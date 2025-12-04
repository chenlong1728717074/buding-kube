package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"sort"
	"sync"
)

var (
	kubeSrv     *KubeSrvService
	kubeSrvOnce sync.Once
)

type KubeSrvService struct {
}

func NewKubeSrvService() *KubeSrvService {
	return &KubeSrvService{}
}

func GetSingletonKubeSrvService() *KubeSrvService {
	kubeSrvOnce.Do(func() {
		kubeSrv = NewKubeSrvService()
	})
	return kubeSrv
}

func (s *KubeSrvService) List(query dto.ServiceQueryDTO) ([]vo.ServiceVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().Services(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取service失败: %v", err)
		return nil, fmt.Errorf("获取service失败: %v", err)
	}
	result := make([]vo.ServiceVO, 0)
	for _, item := range items.Items {
		vi := vo.Service2VO(item)
		copy := item.DeepCopy()
		copy.ObjectMeta.ManagedFields = nil
		copy.ObjectMeta.ResourceVersion = ""
		copy.ObjectMeta.CreationTimestamp = metav1.Time{}
		copy.Status = corev1.ServiceStatus{}
		yamlData, err := yaml.Marshal(copy)
		if err != nil {
			logs.Error("序列化Service失败: %v", err)
		}
		vi.Yaml = string(yamlData)
		result = append(result, vi)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

func (s *KubeSrvService) GetInfo(query dto.ServiceBaseDTO) (*vo.ServiceVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	item, err := clientSet.CoreV1().Services(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取service详情失败: %v", err)
		return nil, fmt.Errorf("获取service详情失败: %v", err)
	}
	vi := vo.Service2VO(*item)
	copy := item.DeepCopy()
	copy.ObjectMeta.ManagedFields = nil
	copy.ObjectMeta.ResourceVersion = ""
	copy.ObjectMeta.CreationTimestamp = metav1.Time{}
	copy.Status = corev1.ServiceStatus{}
	yamlData, err := yaml.Marshal(copy)
	if err != nil {
		logs.Error("序列化Service失败: %v", err)
	}
	vi.Yaml = string(yamlData)
	return &vi, nil
}

func (s *KubeSrvService) GetYaml(query dto.ServiceBaseDTO) (string, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return "", errors.New("获取集群失败")
	}
	item, err := clientSet.CoreV1().Services(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取service失败: %v", err)
		return "", fmt.Errorf("获取service失败: %v", err)
	}
	copy := item.DeepCopy()
	copy.ObjectMeta.ManagedFields = nil
	copy.ObjectMeta.ResourceVersion = ""
	copy.ObjectMeta.CreationTimestamp = metav1.Time{}
	copy.Status = corev1.ServiceStatus{}
	yamlData, err := yaml.Marshal(copy)
	if err != nil {
		logs.Error("序列化Service失败: %v", err)
		return "", fmt.Errorf("序列化Service失败: %v", err)
	}
	return string(yamlData), nil
}
