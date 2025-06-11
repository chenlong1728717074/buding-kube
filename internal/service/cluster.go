package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/pkg/kube"
	"buding-kube/pkg/utils"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sync"
)

var (
	clusterSrv  *ClusterService
	clusterOnce sync.Once
	ClusterMap  map[string]*kubernetes.Clientset
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
	status, clientSet, err := s.getClusterStatus(create)
	if err != nil {
		return errors.New("获取集群状态失败," + err.Error())
	}
	var clusterConfigSecret corev1.Secret
	clusterConfigSecret.Name = create.Id
	//Labels
	clusterConfigSecret.Labels = make(map[string]string)
	clusterConfigSecret.Labels["status"] = "true"
	//Annotations
	clusterConfigSecret.Annotations = make(map[string]string)
	clusterConfigSecret.Annotations = utils.Struct2Map(status)
	//StringData
	clusterConfigSecret.StringData = make(map[string]string)
	clusterConfigSecret.StringData["kubeconfig"] = create.Config

	if clusterConfigSecret.Name == "" {
		clusterConfigSecret.Name = uuid.New().String()
		_, err = kube.InClusterClientSet.CoreV1().Secrets("buding").
			Create(context.TODO(), &clusterConfigSecret, metav1.CreateOptions{})
		return err
	}
	_, err = kube.InClusterClientSet.CoreV1().Secrets("buding").Update(context.TODO(), &clusterConfigSecret, metav1.UpdateOptions{})
	//全局clientSet
	ClusterMap[create.Id] = clientSet
	return err
}

func (s *ClusterService) getClusterStatus(create dto.NodeCreateDTO) (*ClusterStatus, *kubernetes.Clientset, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(create.Config))
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, nil, err
	}
	clusterVersion := serverVersion.String()
	var result ClusterStatus
	copier.Copy(&result, &create)
	result.Status = "Active"
	result.Version = clusterVersion
	return &result, clientset, nil
}

func (s *ClusterService) Delete(id string) error {
	err := kube.InClusterClientSet.CoreV1().Secrets("buding").Delete(context.TODO(), id,
		metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	delete(ClusterMap, id)
	return nil
}
