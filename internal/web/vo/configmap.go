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
	DataCount         int               `json:"dataCount"` // 数据条数(列表接口不返回 data 本身)
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

// ConfigMapSummary2VO 只根据元数据构造列表项, 不包含 data / labels 等大字段
func ConfigMapSummary2VO(meta metav1.ObjectMeta) ConfigMapVO {
	return ConfigMapVO{
		BaseVO: BaseVO{
			Name:      meta.Name,
			Namespace: meta.Namespace,
			Alias:     meta.Annotations["alias"],
			Describe:  meta.Annotations["describe"],
		},
		CreationTimestamp: meta.CreationTimestamp.Time,
		CreateTime:        meta.CreationTimestamp.Time,
	}
}
