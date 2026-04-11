package service

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

func (s *ClusterService) SaveOrUpdate(create dto.NodeUpdateDTO) error {
	name := strings.TrimSpace(create.Name)
	if name == "" {
		return errors.New("集群名称不能为空")
	}

	existing, err := kube.GetCluster(name)
	if err != nil && !apierrors.IsNotFound(err) {
		logs.Error("查询集群失败: %v", err)
		return err
	}
	if err != nil && apierrors.IsNotFound(err) {
		existing = nil
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
		obj.ResourceVersion = existing.ResourceVersion
		if err = kube.UpdateCluster(obj); err != nil {
			logs.Error("更新集群失败: %v", err)
			return err
		}
	}

	ClusterMap.Put(name, clientSet, restCfg)
	return nil
}

func (s *ClusterService) resolveClusterCredential(create dto.NodeUpdateDTO, existing *kube.Cluster) (string, string, error) {
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
		return buildKubeConfigFromToken(uri, token), "token", nil
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

func buildKubeConfigFromToken(uri, token string) string {
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		uri = "https://" + uri
	}
	return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: cluster
  cluster:
    server: %s
    insecure-skip-tls-verify: true
users:
- name: user
  user:
    token: %s
contexts:
- name: ctx
  context:
    cluster: cluster
    user: user
current-context: ctx
`, uri, token)
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

func (s *ClusterService) List(query dto.PageQueryDTO) ([]vo.ClusterQueryVO, error) {
	clusters, err := kube.ListClusters(metav1.ListOptions{})
	if err != nil {
		logs.Error("获取集群资源失败%v", err)
		return nil, err
	}
	result := make([]vo.ClusterQueryVO, 0)
	keyword := strings.TrimSpace(query.Keyword)
	for _, item := range clusters {
		status := item.Status.State
		if status == "" {
			switch strings.ToLower(item.Status.Health) {
			case "healthy":
				status = "Active"
			case "unhealthy":
				status = "Error"
			default:
				status = "Unknown"
			}
		}
		version := item.Status.Version
		if version == "" {
			version = "-"
		}
		name := item.Name
		alias := item.Spec.Alias
		if keyword != "" && !strings.Contains(name, keyword) && !strings.Contains(alias, keyword) {
			continue
		}
		result = append(result, vo.ClusterQueryVO{
			Id:        name,
			Name:      name,
			Alias:     alias,
			Describe:  item.Spec.Description,
			Status:    status,
			Version:   version,
			ApiServer: item.Spec.ApiServer,
			Endpoint:  item.Spec.ApiServer,
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
	status := item.Status.State
	if status == "" {
		switch strings.ToLower(item.Status.Health) {
		case "healthy":
			status = "Active"
		case "unhealthy":
			status = "Error"
		default:
			status = "Unknown"
		}
	}
	version := item.Status.Version
	if version == "" {
		version = "-"
	}
	decryptedConfig, err := utils.DecryptSensitive(item.Spec.KubeConfig)
	if err != nil {
		return nil, err
	}
	return &vo.ClusterVO{
		Id:        item.Name,
		Name:      item.Name,
		Alias:     item.Spec.Alias,
		Describe:  item.Spec.Description,
		Status:    status,
		Version:   version,
		Config:    decryptedConfig,
		ApiServer: item.Spec.ApiServer,
		Endpoint:  item.Spec.ApiServer,
	}, nil
}
