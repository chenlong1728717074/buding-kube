package vo

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

// EndpointVO 已过时，建议使用 EndpointSliceVO
// Deprecated: Use EndpointSliceVO instead
type EndpointVO struct {
	Name       string    `json:"name"`       // 端点资源的名称
	Namespace  string    `json:"namespace"`  // 端点资源所属的命名空间
	CreateTime time.Time `json:"createTime"` // 端点资源的创建时间
	Yaml       string    `json:"yaml"`       // 端点资源的YAML配置内容
}

// Endpoint2VO 已过时，建议使用 EndpointSlice2VO
// Deprecated: Use EndpointSlice2VO instead
func Endpoint2VO(ep corev1.Endpoints) EndpointVO {
	return EndpointVO{
		Name:       ep.Name,
		Namespace:  ep.Namespace,
		CreateTime: ep.CreationTimestamp.Time,
	}
}
