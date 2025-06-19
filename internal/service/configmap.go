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
	configMapSrv  *ConfigMapService
	configMapOnce sync.Once
)

type ConfigMapService struct {
}

func NewConfigMapService() *ConfigMapService {
	return &ConfigMapService{}
}

func GetSingletonConfigMapService() *ConfigMapService {
	configMapOnce.Do(func() {
		configMapSrv = NewConfigMapService()
	})
	return configMapSrv
}

func (s *ConfigMapService) List(query dto.ConfigMapPageQueryBaseDTO) ([]vo.ConfigMapVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().ConfigMaps(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取ConfigMap失败: %v", err)
		return nil, fmt.Errorf("获取ConfigMap失败: %v", err)
	}
	result := make([]vo.ConfigMapVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.ConfigMap2VO(item)
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化ConfigMap失败: %v", err)
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
