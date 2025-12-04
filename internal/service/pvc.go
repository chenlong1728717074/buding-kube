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
	"strings"
	"sync"
)

var (
	pvcSrv  *PVCService
	pvcOnce sync.Once
)

type PVCService struct {
}

func NewPVCService() *PVCService {
	return &PVCService{}
}

func GetSingletonPVCService() *PVCService {
	pvcOnce.Do(func() {
		pvcSrv = NewPVCService()
	})
	return pvcSrv
}

func (s *PVCService) List(query dto.PVCPageQueryBaseDTO) ([]vo.PVCVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().PersistentVolumeClaims(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取PVC失败: %v", err)
		return nil, fmt.Errorf("获取PVC失败: %v", err)
	}
	result := make([]vo.PVCVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.PVC2VO(item)
			copy := item.DeepCopy()
			copy.ObjectMeta.ManagedFields = nil
			copy.ObjectMeta.ResourceVersion = ""
			copy.ObjectMeta.CreationTimestamp = metav1.Time{}
			copy.Status = corev1.PersistentVolumeClaimStatus{}
			yamlData, err := yaml.Marshal(copy)
			if err != nil {
				logs.Error("序列化PVC失败: %v", err)
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
