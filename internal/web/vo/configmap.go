package vo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type ConfigMapVO struct {
	BaseVO
	Data              map[string]string `json:"data"`
	BinaryData        map[string][]byte `json:"binaryData"`
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

func ConfigMap2VO(cm corev1.ConfigMap) ConfigMapVO {
	vo := ConfigMapVO{
		BaseVO: BaseVO{
			Name:      cm.Name,
			Namespace: cm.Namespace,
			Alias:     cm.Annotations["alias"],
			Describe:  cm.Annotations["describe"],
		},
		Data:              cm.Data,
		BinaryData:        cm.BinaryData,
		Generation:        cm.Generation,
		Version:           cm.APIVersion,
		Uid:               string(cm.UID),
		ResourceVersion:   cm.ResourceVersion,
		Annotations:       cm.Annotations,
		Labels:            cm.Labels,
		CreationTimestamp: cm.CreationTimestamp.Time,
		CreateTime:        cm.CreationTimestamp.Time,
		Metadata:          cm.ObjectMeta,
	}
	return vo
}
