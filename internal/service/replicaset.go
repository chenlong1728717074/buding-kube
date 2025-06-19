package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
	"sync"
)

var (
	replicaSetSrv  *ReplicaSetService
	replicaSetOnce sync.Once
)

type ReplicaSetService struct {
}

func NewReplicaSetService() *ReplicaSetService {
	return &ReplicaSetService{}
}

func GetSingletonReplicaSetService() *ReplicaSetService {
	replicaSetOnce.Do(func() {
		replicaSetSrv = NewReplicaSetService()
	})
	return replicaSetSrv
}

// List 获取ReplicaSet列表
func (s *ReplicaSetService) List(query dto.WorkloadQueryDTO) ([]vo.WorkloadVO, error) {
	client, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}

	replicaSets, err := client.AppsV1().ReplicaSets(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取ReplicaSet列表失败: %v", err)
		return nil, err
	}

	var result []vo.WorkloadVO
	for _, rs := range replicaSets.Items {
		// 根据关键字过滤
		if query.Name != "" && !strings.Contains(rs.Name, query.Name) {
			continue
		}

		workloadVO := vo.ReplicaSet2VO(&rs)
		result = append(result, workloadVO)
	}

	// 按创建时间倒序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})

	return result, nil
}
