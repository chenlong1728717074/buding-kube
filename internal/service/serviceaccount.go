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
	serviceAccountSrv  *ServiceAccountService
	serviceAccountOnce sync.Once
)

type ServiceAccountService struct {
}

func NewServiceAccountService() *ServiceAccountService {
	return &ServiceAccountService{}
}

func GetSingletonServiceAccountService() *ServiceAccountService {
	serviceAccountOnce.Do(func() {
		serviceAccountSrv = NewServiceAccountService()
	})
	return serviceAccountSrv
}

func (s *ServiceAccountService) List(query dto.ServiceAccountPageQueryBaseDTO) ([]vo.ServiceAccountVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().ServiceAccounts(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取ServiceAccount失败: %v", err)
		return nil, fmt.Errorf("获取ServiceAccount失败: %v", err)
	}
	result := make([]vo.ServiceAccountVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.ServiceAccount2VO(item)
			copy := item.DeepCopy()
			copy.ObjectMeta.ManagedFields = nil
			copy.ObjectMeta.ResourceVersion = ""
			copy.ObjectMeta.CreationTimestamp = metav1.Time{}
			yamlData, err := yaml.Marshal(copy)
			if err != nil {
				logs.Error("序列化ServiceAccount失败: %v", err)
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
