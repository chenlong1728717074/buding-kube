package vo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type PVCVO struct {
	BaseVO
	AccessModes       []corev1.PersistentVolumeAccessMode     `json:"accessModes"`
	Resources         corev1.VolumeResourceRequirements       `json:"resources"`
	StorageClassName  *string                                 `json:"storageClassName"`
	VolumeMode        *corev1.PersistentVolumeMode            `json:"volumeMode"`
	VolumeName        string                                  `json:"volumeName"`
	Phase             corev1.PersistentVolumeClaimPhase       `json:"phase"`
	Capacity          corev1.ResourceList                     `json:"capacity"`
	Conditions        []corev1.PersistentVolumeClaimCondition `json:"conditions"`
	Generation        int64                                   `json:"generation"`
	Version           string                                  `json:"version"`
	Uid               string                                  `json:"uid"`
	ResourceVersion   string                                  `json:"resourceVersion"`
	Annotations       map[string]string                       `json:"annotations"`
	Labels            map[string]string                       `json:"labels"`
	CreationTimestamp time.Time                               `json:"creationTimestamp"`
	CreateTime        time.Time                               `json:"createTime"`
	Metadata          metav1.ObjectMeta                       `json:"metadata"`
	Yaml              string                                  `json:"yaml"`
}

func PVC2VO(pvc corev1.PersistentVolumeClaim) PVCVO {
	return PVCVO{
		BaseVO: BaseVO{
			Name:      pvc.Name,
			Namespace: pvc.Namespace,
		},
		AccessModes:       pvc.Spec.AccessModes,
		Resources:         pvc.Spec.Resources,
		StorageClassName:  pvc.Spec.StorageClassName,
		VolumeMode:        pvc.Spec.VolumeMode,
		VolumeName:        pvc.Spec.VolumeName,
		Phase:             pvc.Status.Phase,
		Capacity:          pvc.Status.Capacity,
		Conditions:        pvc.Status.Conditions,
		Generation:        pvc.Generation,
		Version:           pvc.APIVersion,
		Uid:               string(pvc.UID),
		ResourceVersion:   pvc.ResourceVersion,
		Annotations:       pvc.Annotations,
		Labels:            pvc.Labels,
		CreationTimestamp: pvc.CreationTimestamp.Time,
		CreateTime:        pvc.CreationTimestamp.Time,
		Metadata:          pvc.ObjectMeta,
	}
}
