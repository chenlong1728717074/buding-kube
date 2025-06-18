package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
)

var (
	deploymentSrv  *DeploymentService
	deploymentOnce sync.Once
)

type DeploymentService struct {
}

func NewDeploymentService() *DeploymentService {
	return &DeploymentService{}
}

func GetSingletonDeploymentService() *DeploymentService {
	deploymentOnce.Do(func() {
		deploymentSrv = NewDeploymentService()
	})
	return deploymentSrv
}

func (s *DeploymentService) List(query dto.DeploymentQueryDTO) ([]vo.DeploymentVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	listOptions := metav1.ListOptions{}
	dyts, err := clientSet.AppsV1().Deployments(query.Namespace).List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取Deployments失败: %v", err)
		return nil, err
	}
	result := make([]vo.DeploymentVO, 0)
	for _, item := range dyts.Items {
		if query.Name == "" || strings.Contains(item.Name, query.Name) {
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化Deployment失败: %v", err)
			}
			dv := vo.Deployment2VO(item)
			dv.Yaml = string(yamlData)
			result = append(result, dv)
		}
	}
	return result, nil
}
