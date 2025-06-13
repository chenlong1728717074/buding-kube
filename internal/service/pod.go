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
	podSrv  *PodService
	podOnce sync.Once
)

type PodService struct {
}

func GetSingletonPodService() *PodService {
	podOnce.Do(func() {
		podSrv = NewPodService()
	})
	return podSrv
}

func NewPodService() *PodService {
	return &PodService{}
}

func (s *PodService) Delete(pod dto.PodDTO) error {
	clientSet, err := ClusterMap.Get(pod.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", pod.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	options := metav1.DeleteOptions{}
	//强制删除
	if pod.Force {
		flag := int64(0)
		options.GracePeriodSeconds = &flag
	}
	err = clientSet.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除pod失败: %s %s", pod.ClusterId, err.Error())
		return err
	}
	return nil
}

func (s *PodService) List(query dto.PodQueryDTO) ([]vo.PodVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	listOptions := metav1.ListOptions{}
	if query.Status != "" {
		listOptions.FieldSelector = fmt.Sprintf("status.phase=%s", query.Status)
	}
	namespaces, err := clientSet.CoreV1().Pods(query.Namespace).List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取pod失败: %v", err)
		return nil, err
	}
	result := make([]vo.PodVO, 0)
	for _, item := range namespaces.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			result = append(result, vo.Pod2VO(item))
		}
	}
	//按照时间倒叙排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTimestamp.After(result[j].CreationTimestamp)
	})
	return result, nil
}

func (s *PodService) GetById(query dto.PodDTO) (*vo.PodInfoVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	pod, err := clientSet.CoreV1().Pods(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取命pod失败: %v", err)
		return nil, err
	}
	yamlData, err := yaml.Marshal(pod)
	if err != nil {
		logs.Error("序列化pod失败: %v", err)
		return nil, err
	}
	events, err := clientSet.CoreV1().Events(query.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", query.Name),
	})
	if err != nil {
		logs.Error("获取pod事件: %v", err)
		return nil, err
	}
	result := vo.Pod2InfoVO(pod, events)
	result.Yaml = string(yamlData)
	return &result, nil
}
