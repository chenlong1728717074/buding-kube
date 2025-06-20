package dto

type ServiceQueryDTO struct {
	PageQueryDTO
	NamespaceBaseDTO
}

type ServiceBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace" binding:"required"`
	Name      string `json:"name" form:"name" binding:"required"`
}
