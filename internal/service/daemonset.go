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
	daemonSetSrv  *DaemonSetService
	daemonSetOnce sync.Once
)

type DaemonSetService struct {
}

func NewDaemonSetService() *DaemonSetService {
	return &DaemonSetService{}
}

func GetSingletonDaemonSetService() *DaemonSetService {
	daemonSetOnce.Do(func() {
		daemonSetSrv = NewDaemonSetService()
	})
	return daemonSetSrv
}

func (s *DaemonSetService) List(query dto.WorkloadQueryDTO) ([]vo.WorkloadVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	listOptions := metav1.ListOptions{}
	// 支持空命名空间查询所有命名空间
	namespace := query.Namespace
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	ds, err := clientSet.AppsV1().DaemonSets(namespace).List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取DaemonSets失败: %v", err)
		return nil, err
	}
	result := make([]vo.WorkloadVO, 0)
	for _, item := range ds.Items {
		if query.Name == "" || strings.Contains(item.Name, query.Name) {
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化DaemonSet失败: %v", err)
			}
			dv := vo.DaemonSet2WorkloadVO(item)
			dv.Yaml = string(yamlData)
			result = append(result, dv)
		}
	}
	return result, nil
}
