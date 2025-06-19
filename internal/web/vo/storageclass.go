package vo

import (
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type StorageClassVO struct {
	Name                 string                            `json:"name"`
	Provisioner          string                            `json:"provisioner"`
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
	return StorageClassVO{
		Name:                 sc.Name,
		Provisioner:          sc.Provisioner,
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
