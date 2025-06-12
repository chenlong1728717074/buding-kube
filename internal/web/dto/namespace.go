package dto

type NamespaceCreateDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace" binding:"required"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
}

type NamespaceBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace" binding:"required"`
	Force     bool   `json:"force" form:"force"`
}

type NamespacePageQueryBaseDTO struct {
	PageQueryDTO
	NamespaceBaseDTO
}
