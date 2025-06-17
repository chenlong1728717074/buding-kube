package dto

type PodDTO struct {
	NamespaceBaseDTO
	Name string `json:"name" form:"name" binding:"required"`
}

type PodQueryDTO struct {
	NamespaceBaseDTO
	PageQueryDTO
	Status string `json:"status" form:"status"`
}

type PodLogDTO struct {
	PodDTO
	SinceTime *CustomTime `json:"sinceTime"`
	Follow    bool        `json:"follow"`
	TailLines *int64      `json:"tailLines"`
}
