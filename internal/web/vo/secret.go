package vo

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	DataCount         int               `json:"dataCount"` // 数据条数, 列表接口使用(列表不返回 data 本身)
}

func Secret2VO(secret corev1.Secret) SecretVO {
	alias := ""
	describe := ""
	if secret.Annotations != nil {
		alias = secret.Annotations["alias"]
		describe = secret.Annotations["describe"]
	}
	return SecretVO{
		BaseVO: BaseVO{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Alias:     alias,
			Describe:  describe,
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

// SecretSummary2VO 只根据元数据构造列表项, 不包含 data / labels 等大字段
func SecretSummary2VO(meta metav1.ObjectMeta) SecretVO {
	alias := ""
	describe := ""
	if meta.Annotations != nil {
		alias = meta.Annotations["alias"]
		describe = meta.Annotations["describe"]
	}
	return SecretVO{
		BaseVO: BaseVO{
			Name:      meta.Name,
			Namespace: meta.Namespace,
			Alias:     alias,
			Describe:  describe,
		},
		CreationTimestamp: meta.CreationTimestamp.Time,
		CreateTime:        meta.CreationTimestamp.Time,
	}
}
