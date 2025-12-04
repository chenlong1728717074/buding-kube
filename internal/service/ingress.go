package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
	"sync"
)

var (
	ingressSrv  *IngressService
	ingressOnce sync.Once
)

type IngressService struct {
}

func NewIngressService() *IngressService {
	return &IngressService{}
}

func GetSingletonIngressService() *IngressService {
	ingressOnce.Do(func() {
		ingressSrv = NewIngressService()
	})
	return ingressSrv
}

func (s *IngressService) List(query dto.IngressPageQueryBaseDTO) ([]vo.IngressVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.NetworkingV1().Ingresses(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取Ingress失败: %v", err)
		return nil, fmt.Errorf("获取Ingress失败: %v", err)
	}
	result := make([]vo.IngressVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.Ingress2VO(item)
			copy := item.DeepCopy()
			copy.ObjectMeta.ManagedFields = nil
			copy.ObjectMeta.ResourceVersion = ""
			copy.ObjectMeta.CreationTimestamp = metav1.Time{}
			copy.Status = networkingv1.IngressStatus{}
			yamlData, err := yaml.Marshal(copy)
			if err != nil {
				logs.Error("序列化Ingress失败: %v", err)
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
