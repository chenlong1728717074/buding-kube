package vo

import (
	networkingv1 "k8s.io/api/networking/v1"
	"time"
)

type IngressVO struct {
	BaseVO
	Alias            string    `json:"alias"`            // 资源别名
	Describe         string    `json:"describe"`         // 资源描述信息
	IngressClassName *string   `json:"ingressClassName"` // Ingress类名称，指定使用的Ingress控制器
	CreateTime       time.Time `json:"createTime"`       // 资源创建时间
	Yaml             string    `json:"yaml"`             // 资源的YAML配置内容
}

func Ingress2VO(ing networkingv1.Ingress) IngressVO {
	return IngressVO{
		BaseVO: BaseVO{
			Name:      ing.Name,
			Namespace: ing.Namespace,
		},
		Alias:            ing.Annotations["alias"],
		Describe:         ing.Annotations["describe"],
		IngressClassName: ing.Spec.IngressClassName,
		CreateTime:       ing.CreationTimestamp.Time,
	}
}
