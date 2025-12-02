package vo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type PVVO struct {
	BaseVO
	Capacity                      corev1.ResourceList                  `json:"capacity"`
	AccessModes                   []corev1.PersistentVolumeAccessMode  `json:"accessModes"`
	PersistentVolumeReclaimPolicy corev1.PersistentVolumeReclaimPolicy `json:"persistentVolumeReclaimPolicy"`
	StorageClassName              string                               `json:"storageClassName"`
	VolumeMode                    *corev1.PersistentVolumeMode         `json:"volumeMode"`
	Phase                         corev1.PersistentVolumePhase         `json:"phase"`
	Message                       string                               `json:"message"`
	Reason                        string                               `json:"reason"`
	Generation                    int64                                `json:"generation"`
	Version                       string                               `json:"version"`
	Uid                           string                               `json:"uid"`
	ResourceVersion               string                               `json:"resourceVersion"`
	Annotations                   map[string]string                    `json:"annotations"`
	Labels                        map[string]string                    `json:"labels"`
	CreationTimestamp             time.Time                            `json:"creationTimestamp"`
	CreateTime                    time.Time                            `json:"createTime"`
	Metadata                      metav1.ObjectMeta                    `json:"metadata"`
	Yaml                          string                               `json:"yaml"`
}

func PV2VO(pv corev1.PersistentVolume) PVVO {
	return PVVO{
		BaseVO: BaseVO{
			Name:      pv.Name,
			Namespace: pv.Namespace,
		},
		Capacity:                      pv.Spec.Capacity,
		AccessModes:                   pv.Spec.AccessModes,
		PersistentVolumeReclaimPolicy: pv.Spec.PersistentVolumeReclaimPolicy,
		StorageClassName:              pv.Spec.StorageClassName,
		VolumeMode:                    pv.Spec.VolumeMode,
		Phase:                         pv.Status.Phase,
		Message:                       pv.Status.Message,
		Reason:                        pv.Status.Reason,
		Generation:                    pv.Generation,
		Version:                       pv.APIVersion,
		Uid:                           string(pv.UID),
		ResourceVersion:               pv.ResourceVersion,
		Annotations:                   pv.Annotations,
		Labels:                        pv.Labels,
		CreationTimestamp:             pv.CreationTimestamp.Time,
		CreateTime:                    pv.CreationTimestamp.Time,
		Metadata:                      pv.ObjectMeta,
	}
}
