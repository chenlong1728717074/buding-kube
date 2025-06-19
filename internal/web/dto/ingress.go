package dto

type IngressCreateDTO struct {
	ClusterId   string            `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace   string            `json:"namespace" form:"namespace" binding:"required"`
	Name        string            `json:"name" form:"name" binding:"required"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Yaml        string            `json:"yaml"`
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

type IngressApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}
