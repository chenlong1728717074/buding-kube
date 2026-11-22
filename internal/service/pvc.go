package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
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

func (s *PVCService) List(ctx context.Context, query dto.PVCPageQueryBaseDTO) ([]vo.PVCVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().PersistentVolumeClaims(query.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取PVC失败: %v", err)
		return nil, fmt.Errorf("获取PVC失败: %v", err)
	}
	result := make([]vo.PVCVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.PVC2VO(item)
			result = append(result, vi)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

// Delete 删除PersistentVolumeClaim
func (s *PVCService) Delete(ctx context.Context, base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	if err := clientSet.CoreV1().PersistentVolumeClaims(base.Namespace).Delete(ctx, base.Name, metav1.DeleteOptions{}); err != nil {
		logs.Error("删除pvc失败: %v", err)
		return fmt.Errorf("删除pvc失败: %v", err)
	}
	return nil
}

// Create 创建持久卷声明: 通过存储类动态创建, 或绑定一个 Available 状态的已有持久卷
func (s *PVCService) Create(ctx context.Context, req dto.PVCCreateDTO) error {
	clientSet, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		return errors.New("请选择命名空间")
	}
	name := strings.TrimSpace(req.Name)
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		return fmt.Errorf("名称不合法: %s", strings.Join(errs, "; "))
	}
	qty, err := parseStorageQuantity(req.Capacity)
	if err != nil {
		return err
	}
	modes, err := parseAccessModes(req.AccessModes)
	if err != nil {
		return err
	}
	volumeMode, err := parseVolumeMode(req.VolumeMode)
	if err != nil {
		return err
	}

	var storageClassName *string
	if req.StorageClassName != nil {
		v := strings.TrimSpace(*req.StorageClassName)
		storageClassName = &v
	}

	volumeName := strings.TrimSpace(req.VolumeName)
	if volumeName != "" {
		pv, err := clientSet.CoreV1().PersistentVolumes().Get(ctx, volumeName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("持久卷 %s 不存在", volumeName)
			}
			return fmt.Errorf("获取持久卷失败: %v", err)
		}
		if pv.Status.Phase != corev1.VolumeAvailable {
			return fmt.Errorf("持久卷 %s 当前状态为 %s, 只能绑定 Available 状态的持久卷", volumeName, pv.Status.Phase)
		}
		if pvCap, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok && qty.Cmp(pvCap) > 0 {
			return fmt.Errorf("申请容量 %s 超过了持久卷的容量 %s", qty.String(), pvCap.String())
		}
		for _, mode := range modes {
			if !containsAccessMode(pv.Spec.AccessModes, mode) {
				return fmt.Errorf("持久卷 %s 不支持访问模式 %s", volumeName, mode)
			}
		}
		// 存储类必须与持久卷一致才会绑定; 未指定时显式使用持久卷的存储类, 防止被默认存储类接管
		pvClass := pv.Spec.StorageClassName
		if storageClassName == nil {
			storageClassName = &pvClass
		} else if *storageClassName != pvClass {
			return fmt.Errorf("存储类与持久卷不一致(持久卷的存储类为 %q)", pvClass)
		}
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      cleanStringMap(req.Labels),
			Annotations: withAliasDescribe(cleanStringMap(req.Annotations), req.Alias, req.Describe),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: modes,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: qty},
			},
			StorageClassName: storageClassName,
			VolumeName:       volumeName,
			VolumeMode:       &volumeMode,
		},
	}
	if _, err = clientSet.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("持久卷声明 %s/%s 已存在", namespace, name)
		}
		logs.Error("创建PVC失败: %v", err)
		return fmt.Errorf("创建持久卷声明失败: %v", err)
	}
	logs.Info("PVC %s/%s 创建成功", namespace, name)
	return nil
}

func containsAccessMode(modes []corev1.PersistentVolumeAccessMode, target corev1.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == target {
			return true
		}
	}
	return false
}
