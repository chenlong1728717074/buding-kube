package vo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type ServiceAccountVO struct {
	BaseVO
	Secrets                      []corev1.ObjectReference      `json:"secrets"`
	ImagePullSecrets             []corev1.LocalObjectReference `json:"imagePullSecrets"`
	AutomountServiceAccountToken *bool                         `json:"automountServiceAccountToken"`
	Generation                   int64                         `json:"generation"`
	Version                      string                        `json:"version"`
	Uid                          string                        `json:"uid"`
	ResourceVersion              string                        `json:"resourceVersion"`
	Annotations                  map[string]string             `json:"annotations"`
	Labels                       map[string]string             `json:"labels"`
	CreationTimestamp            time.Time                     `json:"creationTimestamp"`
	CreateTime                   time.Time                     `json:"createTime"`
	Metadata                     metav1.ObjectMeta             `json:"metadata"`
	Yaml                         string                        `json:"yaml"`
}

func ServiceAccount2VO(sa corev1.ServiceAccount) ServiceAccountVO {
	return ServiceAccountVO{
		BaseVO: BaseVO{
			Name:      sa.Name,
			Namespace: sa.Namespace,
		},
		Secrets:                      sa.Secrets,
		ImagePullSecrets:             sa.ImagePullSecrets,
		AutomountServiceAccountToken: sa.AutomountServiceAccountToken,
		Generation:                   sa.Generation,
		Version:                      sa.APIVersion,
		Uid:                          string(sa.UID),
		ResourceVersion:              sa.ResourceVersion,
		Annotations:                  sa.Annotations,
		Labels:                       sa.Labels,
		CreationTimestamp:            sa.CreationTimestamp.Time,
		CreateTime:                   sa.CreationTimestamp.Time,
		Metadata:                     sa.ObjectMeta,
	}
}
