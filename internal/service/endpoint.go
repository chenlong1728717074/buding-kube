package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
	"sync"
)

var (
	endpointSrv  *EndpointService
	endpointOnce sync.Once
)

// EndpointService 已过时，建议使用 EndpointSliceService
// Deprecated: Use EndpointSliceService instead
type EndpointService struct {
}

func NewEndpointService() *EndpointService {
	return &EndpointService{}
}

func GetSingletonEndpointService() *EndpointService {
	endpointOnce.Do(func() {
		endpointSrv = NewEndpointService()
	})
	return endpointSrv
}

func (s *EndpointService) List(query dto.EndpointPageQueryBaseDTO) ([]vo.EndpointVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().Endpoints(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取Endpoint失败: %v", err)
		return nil, fmt.Errorf("获取Endpoint失败: %v", err)
	}
	result := make([]vo.EndpointVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.Endpoint2VO(item)
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化Endpoint失败: %v", err)
			}
			vi.Yaml = string(yamlData)
			result = append(result, vi)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}
