package dto

type WorkloadBaseDTO struct {
	NamespaceBaseDTO
	Name string `json:"name" form:"name" `
}

type WorkloadQueryDTO struct {
	PageQueryDTO
	WorkloadBaseDTO
	Status string `json:"status" form:"status"`
}

type WorkloadUpdateDTO struct {
	WorkloadBaseDTO
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
}

type WorkloadApplyDTO struct {
	WorkloadBaseDTO
	Yaml string `json:"yaml"`
}
