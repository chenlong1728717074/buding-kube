package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
	"sync"
)

var (
	cronJobSrv  *CronJobService
	cronJobOnce sync.Once
)

type CronJobService struct {
}

func NewCronJobService() *CronJobService {
	return &CronJobService{}
}

func GetSingletonCronJobService() *CronJobService {
	cronJobOnce.Do(func() {
		cronJobSrv = NewCronJobService()
	})
	return cronJobSrv
}

func (s *CronJobService) List(query dto.CronJobPageQueryBaseDTO) ([]vo.CronJobVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.BatchV1().CronJobs(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取CronJob失败: %v", err)
		return nil, fmt.Errorf("获取CronJob失败: %v", err)
	}
	result := make([]vo.CronJobVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.CronJob2VO(item)
			copy := item.DeepCopy()
			copy.ObjectMeta.ManagedFields = nil
			copy.ObjectMeta.ResourceVersion = ""
			copy.ObjectMeta.CreationTimestamp = metav1.Time{}
			copy.Status = batchv1.CronJobStatus{}
			yamlData, err := yaml.Marshal(copy)
			if err != nil {
				logs.Error("序列化CronJob失败: %v", err)
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
