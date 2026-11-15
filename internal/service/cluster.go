package service

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	clusterSrv  *ClusterService
	clusterOnce sync.Once
	ClusterMap  = NewClusterCacheMap()
)

type ClusterService struct {
}

type ClusterStatus struct {
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
	Version   string `json:"version"`
	Status    string `json:"status"`
	ApiServer string `json:"apiServer"`
}

type ClusterCache struct {
	clientSet *kubernetes.Clientset
	config    *rest.Config
}

type ClusterCacheMap struct {
	caches sync.Map
}

type debugTransport struct {
	rt http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	logs.Info("[HTTP] >>> %s %s", req.Method, req.URL.Path)

	resp, err := t.rt.RoundTrip(req)

	elapsed := time.Since(start)

	if err != nil {
		logs.Error("[HTTP] <<< 失败: %v, 耗时: %v", err, elapsed)
		return resp, err
	}

	logs.Info("[HTTP] <<< status=%d, 耗时: %v", resp.StatusCode, elapsed)
	return resp, err
}

func NewClusterCacheMap() *ClusterCacheMap {
	return &ClusterCacheMap{}
}

func (m *ClusterCacheMap) Put(key string, value *kubernetes.Clientset, restConfig *rest.Config) {
	m.caches.Store(key, &ClusterCache{
		clientSet: value,
		config:    restConfig,
	})
}

func (m *ClusterCacheMap) Delete(key string) {
	m.caches.Delete(key)
}

func (m *ClusterCacheMap) GetClientSet(clusterId string) (*kubernetes.Clientset, error) {
	clientSet, err := ClusterMap.Get(clusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", clusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	return clientSet, nil
}

func (m *ClusterCacheMap) Get(key string) (*kubernetes.Clientset, error) {
	if cache, ok := m.caches.Load(key); ok {
		return cache.(*ClusterCache).clientSet, nil
	}
	cli, _, err := m.InitCache(key)
	return cli, err
}

func (m *ClusterCacheMap) GetConfig(key string) (*rest.Config, error) {
	if cache, ok := m.caches.Load(key); ok {
		return cache.(*ClusterCache).config, nil
	}
	_, config, err := m.InitCache(key)
	return config, err
}

func (m *ClusterCacheMap) GetCache(key string) (*ClusterCache, error) {
	if cache, ok := m.caches.Load(key); ok {
		return cache.(*ClusterCache), nil
	}

	_, _, err := m.InitCache(key)
	if err != nil {
		return nil, err
	}

	cache, ok := m.caches.Load(key)
	if !ok {
		return nil, errors.New("初始化缓存后仍未找到")
	}
	return cache.(*ClusterCache), nil
}

func (m *ClusterCacheMap) InitCache(key string) (*kubernetes.Clientset, *rest.Config, error) {
	if cache, ok := m.caches.Load(key); ok {
		c := cache.(*ClusterCache)
		return c.clientSet, c.config, nil
	}

	cluster, err := kube.GetCluster(key)
	if err != nil {
		logs.Error("获取集群资源失败:%v", err)
		return nil, nil, err
	}
	kubeConfig, err := utils.DecryptSensitive(cluster.Spec.KubeConfig)
	if err != nil {
		return nil, nil, err
	}
	kubeConfig = strings.TrimSpace(kubeConfig)
	if kubeConfig == "" {
		return nil, nil, errors.New("集群 kubeConfig 为空")
	}

	set, restConfig, err := buildClientSet(kubeConfig)
	if err != nil {
		logs.Error("连接到集群资源失败:%v", err)
		return nil, nil, err
	}

	cache := &ClusterCache{
		clientSet: set,
		config:    restConfig,
	}

	actual, loaded := m.caches.LoadOrStore(key, cache)
	if loaded {
		actualCache := actual.(*ClusterCache)
		return actualCache.clientSet, actualCache.config, nil
	}

	return set, restConfig, nil
}

func GetSingletonClusterService() *ClusterService {
	clusterOnce.Do(func() {
		clusterSrv = NewClusterService()
	})
	return clusterSrv
}

func NewClusterService() *ClusterService {
	return &ClusterService{}
}

// clusterNameRegexp k8s 对象名的 DNS-1123 子域格式校验
var clusterNameRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

func validateClusterName(name string) error {
	if len(name) > 253 {
		return errors.New("集群名称过长(最多253个字符)")
	}
	if !clusterNameRegexp.MatchString(name) {
		return errors.New("集群名称只能包含小写字母、数字、'-' 和 '.', 且以字母或数字开头结尾")
	}
	return nil
}

func (s *ClusterService) SaveOrUpdate(create dto.ClusterSaveDTO) error {
	name := strings.TrimSpace(create.Name)
	if name == "" {
		return errors.New("集群名称不能为空")
	}
	if err := validateClusterName(name); err != nil {
		return err
	}

	existing, err := kube.GetCluster(name)
	if err != nil && !apierrors.IsNotFound(err) {
		logs.Error("查询集群失败: %v", err)
		return err
	}

	kubeConfig, authType, err := s.resolveClusterCredential(create, existing)
	if err != nil {
		return err
	}

	status, clientSet, restCfg, err := s.getClusterStatus(name, create.Alias, create.Describe, kubeConfig)
	if err != nil {
		return err
	}

	encryptedConfig, err := utils.EncryptSensitive(kubeConfig)
	if err != nil {
		return err
	}

	obj := kube.BuildCluster(
		name,
		create.Alias,
		create.Describe,
		encryptedConfig,
		authType,
		status.ApiServer,
		status.Version,
		status.Status,
	)

	if existing == nil {
		if err = kube.CreateCluster(obj, kube.GlobalClient.DynamicClient); err != nil {
			logs.Error("创建集群失败: %v", err)
			return err
		}
	} else {
		// 保留不由本接口管理的字段,避免一次普通更新抹掉真实状态
		obj.Status.NodeCount = existing.Status.NodeCount
		obj.Spec.Region = existing.Spec.Region
		obj.Spec.Labels = existing.Spec.Labels
		// ResourceVersion 由 UpdateCluster 内部获取并设置
		if err = kube.UpdateCluster(obj); err != nil {
			logs.Error("更新集群失败: %v", err)
			return err
		}
	}

	ClusterMap.Put(name, clientSet, restCfg)
	return nil
}

func (s *ClusterService) resolveClusterCredential(create dto.ClusterSaveDTO, existing *kube.Cluster) (string, string, error) {
	config := strings.TrimSpace(create.Config)
	uri := strings.TrimSpace(create.Uri)
	token := strings.TrimSpace(create.Token)

	if config != "" {
		return config, "kubeconfig", nil
	}
	if uri != "" || token != "" {
		if uri == "" || token == "" {
			return "", "", errors.New("uri和token必须同时提供")
		}
		config, err := buildKubeConfigFromToken(uri, token)
		if err != nil {
			return "", "", err
		}
		return config, "token", nil
	}
	if existing == nil {
		return "", "", errors.New("请提供kubeconfig或uri+token")
	}

	existingConfig, err := utils.DecryptSensitive(existing.Spec.KubeConfig)
	if err != nil {
		return "", "", err
	}
	authType := strings.TrimSpace(existing.Spec.Auth.Type)
	if authType == "" {
		authType = "kubeconfig"
	}
	if strings.TrimSpace(existingConfig) == "" {
		return "", "", errors.New("kubeconfig不能为空")
	}
	return existingConfig, authType, nil
}

func buildKubeConfigFromToken(uri, token string) (string, error) {
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		uri = "https://" + uri
	}
	// 结构化构造避免 token 含特殊字符破坏 YAML
	config := clientcmdapi.NewConfig()
	config.Clusters["cluster"] = &clientcmdapi.Cluster{
		Server:                uri,
		InsecureSkipTLSVerify: true,
	}
	config.AuthInfos["user"] = &clientcmdapi.AuthInfo{
		Token: token,
	}
	config.Contexts["ctx"] = &clientcmdapi.Context{
		Cluster:  "cluster",
		AuthInfo: "user",
	}
	config.CurrentContext = "ctx"
	data, err := clientcmd.Write(*config)
	if err != nil {
		return "", fmt.Errorf("生成 kubeconfig 失败: %w", err)
	}
	return string(data), nil
}

func (s *ClusterService) getClusterStatus(name, alias, describe, kubeConfig string) (*ClusterStatus, *kubernetes.Clientset, *rest.Config, error) {
	clientset, restConfig, err := buildClientSet(kubeConfig)
	if err != nil {
		return nil, nil, nil, err
	}
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, nil, nil, err
	}
	clusterVersion := serverVersion.String()
	result := &ClusterStatus{
		Name:      name,
		Alias:     alias,
		Describe:  describe,
		Version:   clusterVersion,
		Status:    "Active",
		ApiServer: restConfig.Host,
	}
	return result, clientset, restConfig, nil
}

func buildClientSet(config string) (*kubernetes.Clientset, *rest.Config, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(config))
	if err != nil {
		return nil, nil, err
	}
	restConfig.QPS = 100
	restConfig.Burst = 200
	restConfig.Timeout = 30 * time.Second

	restConfig.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return &debugTransport{rt: rt}
	})
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	return clientSet, restConfig, nil
}

func (s *ClusterService) DeleteByName(name string) error {
	err := kube.DeleteCluster(name)
	if err != nil {
		logs.Error("删除集群失败%s %v", name, err)
		return err
	}
	ClusterMap.Delete(name)
	return nil
}

// displayStatus 集群展示状态:优先取 status.state,旧数据回退到 status.health
func displayStatus(state, health string) string {
	if state != "" {
		return state
	}
	switch strings.ToLower(health) {
	case "healthy":
		return "Active"
	case "unhealthy":
		return "Error"
	default:
		return "Unknown"
	}
}

func displayVersion(version string) string {
	if version == "" {
		return "-"
	}
	return version
}

// ClusterResourceCount 集群实时资源统计
type ClusterResourceCount struct {
	NodeCount       int64
	PodCount        int64
	NamespaceCount  int64
	DeploymentCount int64
}

// countClusterResources 并发统计目标集群的节点/Pod/命名空间/Deployment 数量
func countClusterResources(clientset *kubernetes.Clientset) ClusterResourceCount {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count ClusterResourceCount
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		if nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err == nil {
			count.NodeCount = int64(len(nodes.Items))
		}
	}()

	go func() {
		defer wg.Done()
		// Limit:1 + RemainingItemCount 避免拉取全量 Pod
		if pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{Limit: 1}); err == nil {
			count.PodCount = int64(len(pods.Items))
			if pods.RemainingItemCount != nil {
				count.PodCount += *pods.RemainingItemCount
			}
		}
	}()

	go func() {
		defer wg.Done()
		if namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err == nil {
			count.NamespaceCount = int64(len(namespaces.Items))
		}
	}()

	go func() {
		defer wg.Done()
		// Limit:1 + RemainingItemCount 避免拉取全量 Deployment
		if deployments, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{Limit: 1}); err == nil {
			count.DeploymentCount = int64(len(deployments.Items))
			if deployments.RemainingItemCount != nil {
				count.DeploymentCount += *deployments.RemainingItemCount
			}
		}
	}()

	wg.Wait()
	return count
}

// countForCluster 获取集群 clientset 并统计资源数量, 失败时返回零值不阻塞列表
func (s *ClusterService) countForCluster(name string) ClusterResourceCount {
	clientset, err := ClusterMap.Get(name)
	if err != nil {
		logs.Error("获取集群客户端失败, 资源统计跳过: %s %v", name, err)
		return ClusterResourceCount{}
	}
	return countClusterResources(clientset)
}

func (s *ClusterService) List(query dto.ClusterPageQueryDTO) ([]vo.ClusterQueryVO, error) {
	clusters, err := kube.ListClusters(metav1.ListOptions{})
	if err != nil {
		logs.Error("获取集群资源失败%v", err)
		return nil, err
	}
	result := make([]vo.ClusterQueryVO, 0)
	keyword := strings.TrimSpace(query.Keyword)
	statusFilter := strings.TrimSpace(query.Status)
	for _, item := range clusters {
		status := displayStatus(item.Status.State, item.Status.Health)
		if statusFilter != "" && !strings.EqualFold(status, statusFilter) {
			continue
		}
		name := item.Name
		alias := item.Spec.Alias
		if keyword != "" && !strings.Contains(name, keyword) && !strings.Contains(alias, keyword) {
			continue
		}
		counts := s.countForCluster(name)
		result = append(result, vo.ClusterQueryVO{
			Id:             name,
			Name:           name,
			Alias:          alias,
			Describe:       item.Spec.Description,
			Status:         status,
			Version:        displayVersion(item.Status.Version),
			ApiServer:      item.Spec.ApiServer,
			CreateTime:     item.CreationTimestamp.Format("2006-01-02 15:04:05"),
			NodeCount:      counts.NodeCount,
			PodCount:       counts.PodCount,
			NamespaceCount: counts.NamespaceCount,
		})
	}
	return result, nil
}

func (s *ClusterService) GetByName(name string) (*vo.ClusterVO, error) {
	item, err := kube.GetCluster(name)
	if err != nil {
		logs.Error("获取集群:%v", err)
		return nil, err
	}
	counts := s.countForCluster(name)
	// 注意:不在响应中返回解密后的 kubeconfig,避免任何登录用户拿到目标集群凭证
	return &vo.ClusterVO{
		Id:              item.Name,
		Name:            item.Name,
		Alias:           item.Spec.Alias,
		Describe:        item.Spec.Description,
		Status:          displayStatus(item.Status.State, item.Status.Health),
		Version:         displayVersion(item.Status.Version),
		ApiServer:       item.Spec.ApiServer,
		CreateTime:      item.CreationTimestamp.Format("2006-01-02 15:04:05"),
		NodeCount:       counts.NodeCount,
		PodCount:        counts.PodCount,
		NamespaceCount:  counts.NamespaceCount,
		DeploymentCount: counts.DeploymentCount,
	}, nil
}
