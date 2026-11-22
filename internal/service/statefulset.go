package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	statefulSetSrv  *StatefulSetService
	statefulSetOnce sync.Once
)

// StatefulSetService StatefulSet 列表; 详情、创建、扩缩容、版本回退等操作见 WorkloadService
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

func (s *StatefulSetService) List(query dto.WorkloadQueryDTO) ([]vo.WorkloadVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	// 命名空间为空时查询所有命名空间
	items, err := clientSet.AppsV1().StatefulSets(query.Namespace).List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取StatefulSet失败: %v", err)
		return nil, err
	}
	keyword := strings.TrimSpace(query.Name)
	if keyword == "" {
		keyword = strings.TrimSpace(query.Keyword)
	}
	result := make([]vo.WorkloadVO, 0, len(items.Items))
	for _, item := range items.Items {
		if keyword == "" || strings.Contains(item.Name, keyword) {
			result = append(result, vo.StatefulSet2WorkloadVO(item))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}
