package vo

import (
	discoveryv1 "k8s.io/api/discovery/v1"
	"time"
)

type EndpointSliceVO struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	CreateTime time.Time `json:"createTime"`
	Yaml       string    `json:"yaml"`
}

func EndpointSlice2VO(eps discoveryv1.EndpointSlice) EndpointSliceVO {
	return EndpointSliceVO{
		Name:       eps.Name,
		Namespace:  eps.Namespace,
		CreateTime: eps.CreationTimestamp.Time,
	}
}
