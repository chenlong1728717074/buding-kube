package dto

type NamespaceCreateDTO struct {
	ClusterId   string            `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace   string            `json:"namespace" form:"namespace" binding:"required"`
	Alias       string            `json:"alias"`
	Describe    string            `json:"describe"`
	Annotations map[string]string `json:"annotations"`
}

type NamespaceBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Force     bool   `json:"force" form:"force"`
}

type NamespacePageQueryBaseDTO struct {
	PageQueryDTO
	NamespaceBaseDTO
}
