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
	secretSrv  *SecretService
	secretOnce sync.Once
)

type SecretService struct {
}

func NewSecretService() *SecretService {
	return &SecretService{}
}

func GetSingletonSecretService() *SecretService {
	secretOnce.Do(func() {
		secretSrv = NewSecretService()
	})
	return secretSrv
}

func (s *SecretService) List(query dto.SecretPageQueryBaseDTO) ([]vo.SecretVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().Secrets(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取Secret失败: %v", err)
		return nil, fmt.Errorf("获取Secret失败: %v", err)
	}
	result := make([]vo.SecretVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.Secret2VO(item)
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化Secret失败: %v", err)
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
