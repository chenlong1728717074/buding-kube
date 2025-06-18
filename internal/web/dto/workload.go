package dto

type DeploymentQueryDTO struct {
	PageQueryDTO
	NamespaceBaseDTO
	Name   string `json:"name" form:"name" `
	Status string `json:"status" form:"status"`
}

type StatefulSetQueryDTO struct {
	PageQueryDTO
	NamespaceBaseDTO
	Name   string `json:"name" form:"name" `
	Status string `json:"status" form:"status"`
}

type DaemonSetQueryDTO struct {
	PageQueryDTO
	NamespaceBaseDTO
	Name   string `json:"name" form:"name" `
	Status string `json:"status" form:"status"`
}
