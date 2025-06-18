package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (s *KubeSrvService) List(query dto.ServiceQueryDTO) ([]interface{}, error) {
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
	for _, item := range items.Items {
		fmt.Println(item)
	}
	return nil, nil
}
