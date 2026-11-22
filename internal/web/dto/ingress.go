package dto

// IngressPathDTO 一条转发路径
type IngressPathDTO struct {
	Path string `json:"path"`
	// PathType Prefix(默认) / Exact / ImplementationSpecific
	PathType    string `json:"pathType"`
	ServiceName string `json:"serviceName"`
	// ServicePort 服务端口号或端口名称
	ServicePort string `json:"servicePort"`
}

// IngressRuleDTO 一个域名的规则; Host 为空表示匹配所有域名
type IngressRuleDTO struct {
	Host  string           `json:"host"`
	Paths []IngressPathDTO `json:"paths"`
}

// IngressTLSDTO HTTPS 证书
type IngressTLSDTO struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

// IngressFormDTO 创建 / 编辑应用路由
type IngressFormDTO struct {
	ClusterId        string            `json:"clusterId" binding:"required"`
	Namespace        string            `json:"namespace" binding:"required"`
	Name             string            `json:"name" binding:"required"`
	Alias            string            `json:"alias"`
	Describe         string            `json:"describe"`
	IngressClassName string            `json:"ingressClassName"`
	Rules            []IngressRuleDTO  `json:"rules"`
	TLS              []IngressTLSDTO   `json:"tls"`
	Labels           map[string]string `json:"labels"`
	Annotations      map[string]string `json:"annotations"`
}

type IngressBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type IngressPageQueryBaseDTO struct {
	PageQueryDTO
	IngressBaseDTO
}
