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

// classifyStorageType 按供应者推断存储类型(块存储/对象存储/文件存储/未分类)
func classifyStorageType(provisioner string) string {
	p := strings.ToLower(provisioner)
	switch {
	// 文件存储
	case strings.Contains(p, "nfs"), strings.Contains(p, "cephfs"), strings.Contains(p, "glusterfs"),
		strings.Contains(p, "nas"), strings.Contains(p, "cpfs"), strings.Contains(p, "filestore"),
		strings.Contains(p, "azurefile"), strings.Contains(p, "efs"), strings.Contains(p, "juicefs"),
		strings.Contains(p, "jfs"):
		return "文件存储"
	// 对象存储
	case strings.Contains(p, "oss"), strings.Contains(p, "s3"), strings.Contains(p, "obs"),
		strings.Contains(p, "cos"), strings.Contains(p, "minio"), strings.Contains(p, "object"),
		strings.Contains(p, "bos"):
		return "对象存储"
	// 块存储
	case strings.Contains(p, "disk"), strings.Contains(p, "rbd"), strings.Contains(p, "block"),
		strings.Contains(p, "ebs"), strings.Contains(p, "longhorn"), strings.Contains(p, "openebs"),
		strings.Contains(p, "rook"), strings.Contains(p, "cbs"), strings.Contains(p, "csi"),
		strings.Contains(p, "local"):
		return "块存储"
	default:
		return "未分类"
	}
}

// provisionerSupportsClone 基于供应者推断是否支持卷克隆(常见 CSI/存储驱动支持 DataSource 克隆)
func provisionerSupportsClone(provisioner string) bool {
	p := strings.ToLower(provisioner)
	switch {
	case strings.Contains(p, "csi"), strings.Contains(p, "rbd"), strings.Contains(p, "rook"),
		strings.Contains(p, "ceph"), strings.Contains(p, "ebs"), strings.Contains(p, "longhorn"),
		strings.Contains(p, "openebs"):
		return true
	default:
		return false
	}
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

	// 统计每个存储类关联的持久卷声明(PVC)数量
	pvcCountMap := make(map[string]int64)
	if pvcs, err := clientSet.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{}); err == nil {
		for _, pvc := range pvcs.Items {
			if pvc.Spec.StorageClassName == nil {
				// 未显式指定时使用集群默认存储类, 不计入具体存储类
				continue
			}
			pvcCountMap[*pvc.Spec.StorageClassName]++
		}
	} else {
		logs.Error("获取PVC列表失败, 持久卷声明数量统计跳过: %v", err)
	}

	result := make([]vo.StorageClassVO, 0)
	for _, item := range items.Items {
		storageType := classifyStorageType(item.Provisioner)
		if query.StorageType != "" && storageType != query.StorageType {
			continue
		}
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.StorageClass2VO(item)
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化StorageClass失败: %v", err)
			}
			vi.Yaml = string(yamlData)
			vi.StorageType = storageType
			vi.PVCCount = pvcCountMap[item.Name]
			vi.AllowVolumeClone = provisionerSupportsClone(item.Provisioner)
			result = append(result, vi)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

// Delete 删除StorageClass
func (s *StorageClassService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	if err := clientSet.StorageV1().StorageClasses().Delete(context.TODO(), base.Name, metav1.DeleteOptions{}); err != nil {
		logs.Error("删除storageclass失败: %v", err)
		return fmt.Errorf("删除storageclass失败: %v", err)
	}
	return nil
}
