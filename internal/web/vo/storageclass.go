package vo

import (
	"buding-kube/pkg/consts"
	"encoding/json"
	"time"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageClassVO struct {
	BaseVO
	Provisioner string `json:"provisioner"`
	// StorageType 存储类型中文名(块存储 / 文件存储 / 对象存储 / 未分类)
	StorageType string `json:"storageType"`
	// StorageTypeKey 存储类型标识(block / file / object, 未分类为空)
	StorageTypeKey string `json:"storageTypeKey"`
	// StorageTypeAuto 为 true 表示存储类型是根据供应者自动识别的, 没有手动设置
	StorageTypeAuto      bool                              `json:"storageTypeAuto"`
	IsDefault            bool                              `json:"isDefault"`
	PVCCount             int64                             `json:"pvcCount"`
	AllowVolumeClone     bool                              `json:"allowVolumeClone"`
	AccessModes          []string                          `json:"accessModes"`
	MountOptions         []string                          `json:"mountOptions"`
	Parameters           map[string]string                 `json:"parameters"`
	ReclaimPolicy        *v1.PersistentVolumeReclaimPolicy `json:"reclaimPolicy"`
	VolumeBindingMode    *storagev1.VolumeBindingMode      `json:"volumeBindingMode"`
	AllowVolumeExpansion *bool                             `json:"allowVolumeExpansion"`
	AllowedTopologies    []v1.TopologySelectorTerm         `json:"allowedTopologies"`
	Generation           int64                             `json:"generation"`
	Version              string                            `json:"version"`
	Uid                  string                            `json:"uid"`
	ResourceVersion      string                            `json:"resourceVersion"`
	Annotations          map[string]string                 `json:"annotations"`
	Labels               map[string]string                 `json:"labels"`
	CreationTimestamp    time.Time                         `json:"creationTimestamp"`
	CreateTime           time.Time                         `json:"createTime"`
	Metadata             metav1.ObjectMeta                 `json:"metadata"`
	Yaml                 string                            `json:"yaml"`
}

func StorageClass2VO(sc storagev1.StorageClass) StorageClassVO {
	ann := sc.Annotations
	alias := firstNonEmpty(ann["alias"], ann[consts.KubeSphereAliasAnnotation])
	describe := firstNonEmpty(ann["describe"], ann[consts.KubeSphereDescriptionAnnotation])
	var accessModes []string
	if raw := ann[consts.StorageClassAccessModesAnnotation]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &accessModes)
	}
	return StorageClassVO{
		BaseVO: BaseVO{
			Name:      sc.Name,
			Namespace: sc.Namespace,
			Alias:     alias,
			Describe:  describe,
		},
		Provisioner:          sc.Provisioner,
		IsDefault:            ann[consts.StorageClassDefaultAnnotation] == "true" || ann[consts.StorageClassDefaultBetaAnnotation] == "true",
		AccessModes:          accessModes,
		MountOptions:         sc.MountOptions,
		Parameters:           sc.Parameters,
		ReclaimPolicy:        sc.ReclaimPolicy,
		VolumeBindingMode:    sc.VolumeBindingMode,
		AllowVolumeExpansion: sc.AllowVolumeExpansion,
		AllowedTopologies:    sc.AllowedTopologies,
		Generation:           sc.Generation,
		Version:              sc.APIVersion,
		Uid:                  string(sc.UID),
		ResourceVersion:      sc.ResourceVersion,
		Annotations:          sc.Annotations,
		Labels:               sc.Labels,
		CreationTimestamp:    sc.CreationTimestamp.Time,
		CreateTime:           sc.CreationTimestamp.Time,
		Metadata:             sc.ObjectMeta,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// StorageClassDetailVO 存储类详情, 附带使用该存储类的持久卷声明
type StorageClassDetailVO struct {
	StorageClassVO
	PersistentVolumeClaims []PVCVO `json:"persistentVolumeClaims"`
	// PVCount 使用该存储类的持久卷数量
	PVCount int64 `json:"pvCount"`
}
