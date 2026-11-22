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
	pvSrv  *PVService
	pvOnce sync.Once
)

type PVService struct {
}

func NewPVService() *PVService {
	return &PVService{}
}

func GetSingletonPVService() *PVService {
	pvOnce.Do(func() {
		pvSrv = NewPVService()
	})
	return pvSrv
}

func (s *PVService) List(query dto.PVPageQueryBaseDTO) ([]vo.PVVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取PV失败: %v", err)
		return nil, fmt.Errorf("获取PV失败: %v", err)
	}
	result := make([]vo.PVVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.PV2VO(item)
			result = append(result, vi)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

// Delete 删除PersistentVolume
func (s *PVService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	if err := clientSet.CoreV1().PersistentVolumes().Delete(context.TODO(), base.Name, metav1.DeleteOptions{}); err != nil {
		logs.Error("删除pv失败: %v", err)
		return fmt.Errorf("删除pv失败: %v", err)
	}
	return nil
}

// Create 创建持久卷
func (s *PVService) Create(ctx context.Context, req dto.PVCreateDTO) error {
	clientSet, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	pv, err := buildPersistentVolume(req)
	if err != nil {
		return err
	}
	if _, err = clientSet.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("持久卷 %s 已存在", pv.Name)
		}
		logs.Error("创建PV失败: %v", err)
		return fmt.Errorf("创建持久卷失败: %v", err)
	}
	logs.Info("PV %s 创建成功", pv.Name)
	return nil
}

var validHostPathTypes = map[corev1.HostPathType]bool{
	corev1.HostPathUnset:             true,
	corev1.HostPathDirectoryOrCreate: true,
	corev1.HostPathDirectory:         true,
	corev1.HostPathFileOrCreate:      true,
	corev1.HostPathFile:              true,
	corev1.HostPathSocket:            true,
	corev1.HostPathCharDev:           true,
	corev1.HostPathBlockDev:          true,
}

func buildPersistentVolume(req dto.PVCreateDTO) (*corev1.PersistentVolume, error) {
	name := strings.TrimSpace(req.Name)
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		return nil, fmt.Errorf("名称不合法: %s", strings.Join(errs, "; "))
	}
	qty, err := parseStorageQuantity(req.Capacity)
	if err != nil {
		return nil, err
	}
	modes, err := parseAccessModes(req.AccessModes)
	if err != nil {
		return nil, err
	}
	volumeMode, err := parseVolumeMode(req.VolumeMode)
	if err != nil {
		return nil, err
	}

	// 静态创建的持久卷默认保留数据, 避免删除 PVC 时误删
	reclaim := corev1.PersistentVolumeReclaimPolicy(strings.TrimSpace(req.ReclaimPolicy))
	switch reclaim {
	case "":
		reclaim = corev1.PersistentVolumeReclaimRetain
	case corev1.PersistentVolumeReclaimRetain, corev1.PersistentVolumeReclaimDelete:
	case corev1.PersistentVolumeReclaimRecycle:
		if req.SourceType != "nfs" && req.SourceType != "hostPath" {
			return nil, errors.New("Recycle 回收策略只支持 NFS 和 HostPath")
		}
	default:
		return nil, fmt.Errorf("不支持的回收策略: %s", reclaim)
	}

	spec := corev1.PersistentVolumeSpec{
		Capacity:                      corev1.ResourceList{corev1.ResourceStorage: qty},
		AccessModes:                   modes,
		PersistentVolumeReclaimPolicy: reclaim,
		StorageClassName:              strings.TrimSpace(req.StorageClassName),
		VolumeMode:                    &volumeMode,
	}
	if opts := cleanStrings(req.MountOptions); len(opts) > 0 {
		spec.MountOptions = opts
	}

	switch req.SourceType {
	case "nfs":
		if req.NFS == nil || strings.TrimSpace(req.NFS.Server) == "" || strings.TrimSpace(req.NFS.Path) == "" {
			return nil, errors.New("请填写 NFS 服务器地址和共享路径")
		}
		spec.NFS = &corev1.NFSVolumeSource{
			Server:   strings.TrimSpace(req.NFS.Server),
			Path:     strings.TrimSpace(req.NFS.Path),
			ReadOnly: req.NFS.ReadOnly,
		}
	case "hostPath":
		if req.HostPath == nil || strings.TrimSpace(req.HostPath.Path) == "" {
			return nil, errors.New("请填写节点目录")
		}
		source := &corev1.HostPathVolumeSource{Path: strings.TrimSpace(req.HostPath.Path)}
		if t := corev1.HostPathType(strings.TrimSpace(req.HostPath.Type)); t != "" {
			if !validHostPathTypes[t] {
				return nil, fmt.Errorf("不支持的目录类型: %s", t)
			}
			source.Type = &t
		}
		spec.HostPath = source
	case "local":
		if req.Local == nil || strings.TrimSpace(req.Local.Path) == "" || strings.TrimSpace(req.Local.Node) == "" {
			return nil, errors.New("本地存储需要填写磁盘路径并选择节点")
		}
		source := &corev1.LocalVolumeSource{Path: strings.TrimSpace(req.Local.Path)}
		if fsType := strings.TrimSpace(req.Local.FSType); fsType != "" {
			source.FSType = &fsType
		}
		spec.Local = source
		// 本地盘只存在于某个节点上, 必须用节点亲和性把使用它的 Pod 调度到该节点
		spec.NodeAffinity = &corev1.VolumeNodeAffinity{
			Required: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "kubernetes.io/hostname",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{strings.TrimSpace(req.Local.Node)},
					}},
				}},
			},
		}
	case "csi":
		if req.CSI == nil || strings.TrimSpace(req.CSI.Driver) == "" || strings.TrimSpace(req.CSI.VolumeHandle) == "" {
			return nil, errors.New("请填写 CSI 驱动名称和卷 ID")
		}
		spec.CSI = &corev1.CSIPersistentVolumeSource{
			Driver:           strings.TrimSpace(req.CSI.Driver),
			VolumeHandle:     strings.TrimSpace(req.CSI.VolumeHandle),
			FSType:           strings.TrimSpace(req.CSI.FSType),
			ReadOnly:         req.CSI.ReadOnly,
			VolumeAttributes: cleanStringMap(req.CSI.VolumeAttributes),
		}
	default:
		return nil, fmt.Errorf("不支持的存储来源: %s, 可选 nfs / hostPath / local / csi", req.SourceType)
	}
	if volumeMode == corev1.PersistentVolumeBlock && spec.Local == nil && spec.CSI == nil {
		return nil, errors.New("块设备(Block)模式只支持本地存储和 CSI")
	}

	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      cleanStringMap(req.Labels),
			Annotations: withAliasDescribe(nil, req.Alias, req.Describe),
		},
		Spec: spec,
	}, nil
}
