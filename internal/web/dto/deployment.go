package dto

type DeploymentQueryDTO struct {
	PageQueryDTO
	NamespaceBaseDTO
	Name   string `json:"name" form:"name" `
	Status string `json:"status" form:"status"`
}
