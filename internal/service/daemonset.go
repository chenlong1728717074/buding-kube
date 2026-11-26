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
	daemonSetSrv  *DaemonSetService
	daemonSetOnce sync.Once
)

// DaemonSetService DaemonSet 列表; 详情、创建、扩缩容、版本回退等操作见 WorkloadService
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

func (s *DaemonSetService) List(ctx context.Context, query dto.WorkloadQueryDTO) ([]vo.WorkloadVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	// 命名空间为空时查询所有命名空间（不使用 ResourceVersion:"0" 避免更新后读到过期缓存）
	items, err := clientSet.AppsV1().DaemonSets(query.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logs.Error("获取DaemonSet失败: %v", err)
		return nil, err
	}
	keyword := strings.TrimSpace(query.Name)
	if keyword == "" {
		keyword = strings.TrimSpace(query.Keyword)
	}
	result := make([]vo.WorkloadVO, 0, len(items.Items))
	for _, item := range items.Items {
		if keyword == "" || strings.Contains(item.Name, keyword) {
			result = append(result, vo.DaemonSet2WorkloadVO(item))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}
