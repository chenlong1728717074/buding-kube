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
	storageClassSrv  *StorageClassService
	storageClassOnce sync.Once
)

type StorageClassService struct {
}

func NewStorageClassService() *StorageClassService {
	return &StorageClassService{}
}

func GetSingletonStorageClassService() *StorageClassService {
	storageClassOnce.Do(func() {
		storageClassSrv = NewStorageClassService()
	})
	return storageClassSrv
}

func (s *StorageClassService) List(query dto.StorageClassPageQueryBaseDTO) ([]vo.StorageClassVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取StorageClass失败: %v", err)
		return nil, fmt.Errorf("获取StorageClass失败: %v", err)
	}
	result := make([]vo.StorageClassVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.StorageClass2VO(item)
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化StorageClass失败: %v", err)
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
