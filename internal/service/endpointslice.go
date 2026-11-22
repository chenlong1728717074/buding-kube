package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
	"sync"
)

var (
	endpointSliceSrv  *EndpointSliceService
	endpointSliceOnce sync.Once
)

type EndpointSliceService struct {
}

func NewEndpointSliceService() *EndpointSliceService {
	return &EndpointSliceService{}
}

func GetSingletonEndpointSliceService() *EndpointSliceService {
	endpointSliceOnce.Do(func() {
		endpointSliceSrv = NewEndpointSliceService()
	})
	return endpointSliceSrv
}

func (s *EndpointSliceService) List(ctx context.Context, query dto.EndpointSlicePageQueryBaseDTO) ([]vo.EndpointSliceVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.DiscoveryV1().EndpointSlices(query.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取EndpointSlice失败: %v", err)
		return nil, fmt.Errorf("获取EndpointSlice失败: %v", err)
	}
	result := make([]vo.EndpointSliceVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.EndpointSlice2VO(item)
			result = append(result, vi)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

// Delete 删除EndpointSlice(注意: 由控制器管理, 删除后可能被自动重建)
func (s *EndpointSliceService) Delete(ctx context.Context, base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	if err := clientSet.DiscoveryV1().EndpointSlices(base.Namespace).Delete(ctx, base.Name, metav1.DeleteOptions{}); err != nil {
		logs.Error("删除endpointslice失败: %v", err)
		return fmt.Errorf("删除endpointslice失败: %v", err)
	}
	return nil
}
