package vo

import (
	discoveryv1 "k8s.io/api/discovery/v1"
	"time"
)

type EndpointSliceVO struct {
	BaseVO
	AddressType string    `json:"addressType"`
	CreateTime  time.Time `json:"createTime"`
	Yaml        string    `json:"yaml"`
}

func EndpointSlice2VO(eps discoveryv1.EndpointSlice) EndpointSliceVO {
	return EndpointSliceVO{
		BaseVO: BaseVO{
			Name:      eps.Name,
			Namespace: eps.Namespace,
		},
		AddressType: string(eps.AddressType),
		CreateTime:  eps.CreationTimestamp.Time,
	}
}
