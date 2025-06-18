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
	statefulSetSrv  *StatefulSetService
	statefulSetOnce sync.Once
)

type StatefulSetService struct {
}

func NewStatefulSetService() *StatefulSetService {
	return &StatefulSetService{}
}

func GetSingletonStatefulSetService() *StatefulSetService {
	statefulSetOnce.Do(func() {
		statefulSetSrv = NewStatefulSetService()
	})
	return statefulSetSrv
}

func (s *StatefulSetService) List(query dto.StatefulSetQueryDTO) ([]vo.WorkloadVO, error) {
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
	sts, err := clientSet.AppsV1().StatefulSets(namespace).List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取StatefulSets失败: %v", err)
		return nil, err
	}
	result := make([]vo.WorkloadVO, 0)
	for _, item := range sts.Items {
		if query.Name == "" || strings.Contains(item.Name, query.Name) {
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化StatefulSet失败: %v", err)
			}
			sv := vo.StatefulSet2WorkloadVO(item)
			sv.Yaml = string(yamlData)
			result = append(result, sv)
		}
	}
	return result, nil
}
