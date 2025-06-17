package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/kube"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"sync"
)

var (
	clusterSrv                    *ClusterService
	clusterOnce                   sync.Once
	ClusterMap                           = make(ClusterCacheMap)
	ClusterConfigSecretLabelKey   string = "buding-kube.com/cluster.metadata"
	ClusterConfigSecretLabelValue string = "true"
)

type ClusterService struct {
}
type ClusterStatus struct {
	Name     string `json:"name"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Version  string `json:"version"`
	Status   string `json:"status"`
}
type ClusterCacheMap map[string]*ClusterCache

type ClusterCache struct {
	clientSet *kubernetes.Clientset
	config    *rest.Config
}

func (m ClusterCacheMap) Put(key string, value *kubernetes.Clientset, restConfig *rest.Config) {
	m[key] = &ClusterCache{
		clientSet: value,
		config:    restConfig,
	}
}
func (m ClusterCacheMap) Delete(key string) {
	delete(m, key)
}

func (m ClusterCacheMap) Get(key string) (*kubernetes.Clientset, error) {
	cache := m[key]
	if cache != nil {
		return cache.clientSet, nil
	}
	cli, _, err := m.InitCache(key)
	return cli, err
}

func (m ClusterCacheMap) GetConfig(key string) (*rest.Config, error) {
	cache := m[key]
	if cache != nil {
		return cache.config, nil
	}
	_, config, err := m.InitCache(key)
	return config, err
}

func (m ClusterCacheMap) GetCache(key string) (*ClusterCache, error) {
	cache := m[key]
	if cache != nil {
		return cache, nil
	}
	_, _, err := m.InitCache(key)
	return m[key], err
}

func (m ClusterCacheMap) InitCache(key string) (*kubernetes.Clientset, *rest.Config, error) {
	item, err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		Get(context.TODO(), key, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取集群资源失败:%v", err)
		return nil, nil, err
	}
	set, restConfig, err := buildClientSet(string(item.Data["kubeconfig"]))
	if err != nil {
		logs.Error("连接到集群资源失败:%v", err)
		return nil, nil, err
	}
	m.Put(key, set, restConfig)
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

func (s *ClusterService) SaveOrUpdate(create dto.NodeCreateDTO) error {
	status, clientSet, restConfig, err := s.getClusterStatus(create)
	if err != nil {
		logs.Info("获取集群状态失败")
		return errors.New("获取集群状态失败," + err.Error())
	}
	var clusterConfigSecret corev1.Secret
	clusterConfigSecret.Name = create.Id
	//Labels
	clusterConfigSecret.Labels = make(map[string]string)
	clusterConfigSecret.Labels[ClusterConfigSecretLabelKey] = ClusterConfigSecretLabelValue
	//Annotations
	clusterConfigSecret.Annotations = make(map[string]string)
	clusterConfigSecret.Annotations = utils.Struct2Map(status)
	//StringData
	clusterConfigSecret.StringData = make(map[string]string)
	clusterConfigSecret.StringData["kubeconfig"] = create.Config

	if clusterConfigSecret.Name == "" {
		clusterConfigSecret.Name = uuid.New().String()
		if _, err = kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
			Create(context.TODO(), &clusterConfigSecret, metav1.CreateOptions{}); err != nil {
			logs.Error("添加集群失败%v", err)
			return err
		}
		return nil
	}
	if _, err = kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		Update(context.TODO(), &clusterConfigSecret, metav1.UpdateOptions{}); err != nil {
		logs.Info("更新集群失败%s", err)
		return err
	}
	//全局clientSet
	ClusterMap.Put(create.Id, clientSet, restConfig)
	return nil
}

func (s *ClusterService) getClusterStatus(create dto.NodeCreateDTO) (*ClusterStatus, *kubernetes.Clientset, *rest.Config, error) {
	clientset, restConfig, err := buildClientSet(create.Config)
	if err != nil {
		return nil, nil, nil, err
	}
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, nil, nil, err
	}
	clusterVersion := serverVersion.String()
	var result ClusterStatus
	copier.Copy(&result, &create)
	result.Status = "Active"
	result.Version = clusterVersion
	return &result, clientset, restConfig, nil
}

func buildClientSet(config string) (*kubernetes.Clientset, *rest.Config, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(config))
	if err != nil {
		return nil, nil, err
	}
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	return clientSet, restConfig, nil
}

func (s *ClusterService) Delete(id string) error {
	err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).Delete(context.TODO(), id,
		metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除集群失败%s %v", id, err)
		return err
	}
	ClusterMap.Delete(id)
	return nil
}

func (s *ClusterService) List(query dto.PageQueryDTO) ([]vo.ClusterQueryVO, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: ClusterConfigSecretLabelKey + "=" + ClusterConfigSecretLabelValue,
	}
	secretList, err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取集群资源失败%v", err)
		return nil, err
	}
	items := secretList.Items
	result := make([]vo.ClusterQueryVO, 0)
	for _, item := range items {
		var cqv vo.ClusterQueryVO
		if err := utils.Map2Struct(item.Annotations, &cqv); err != nil {
			logs.Error("集群列表转换失败:%v", err)
			return nil, err
		}
		if query.Keyword == "" || strings.Contains(cqv.Name, query.Keyword) {
			cqv.Id = item.Name
			result = append(result, cqv)
		}
	}
	return result, nil
}

func (s *ClusterService) GetById(id string) (*vo.ClusterVO, error) {
	item, err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取集群:%v", err)
		return nil, err
	}
	var cqv vo.ClusterVO
	if err := utils.Map2Struct(item.Annotations, &cqv); err != nil {
		logs.Error("集群信息转换失败:%v", err)
		return nil, err
	}
	cqv.Id = item.Name
	cqv.Config = string(item.Data["kubeconfig"])
	return &cqv, nil
}
