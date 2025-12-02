package vo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type SecretVO struct {
	BaseVO
	Type              corev1.SecretType `json:"type"`
	Data              map[string][]byte `json:"data"`
	StringData        map[string]string `json:"stringData"`
	Immutable         *bool             `json:"immutable"`
	Generation        int64             `json:"generation"`
	Version           string            `json:"version"`
	Uid               string            `json:"uid"`
	ResourceVersion   string            `json:"resourceVersion"`
	Annotations       map[string]string `json:"annotations"`
	Labels            map[string]string `json:"labels"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	CreateTime        time.Time         `json:"createTime"`
	Metadata          metav1.ObjectMeta `json:"metadata"`
	Yaml              string            `json:"yaml"`
}

func Secret2VO(secret corev1.Secret) SecretVO {
	return SecretVO{
		BaseVO: BaseVO{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		},
		Type:              secret.Type,
		Data:              secret.Data,
		StringData:        secret.StringData,
		Immutable:         secret.Immutable,
		Generation:        secret.Generation,
		Version:           secret.APIVersion,
		Uid:               string(secret.UID),
		ResourceVersion:   secret.ResourceVersion,
		Annotations:       secret.Annotations,
		Labels:            secret.Labels,
		CreationTimestamp: secret.CreationTimestamp.Time,
		CreateTime:        secret.CreationTimestamp.Time,
		Metadata:          secret.ObjectMeta,
	}
}
