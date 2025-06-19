package dto

type JobCreateDTO struct {
	ClusterId   string            `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace   string            `json:"namespace" form:"namespace" binding:"required"`
	Name        string            `json:"name" form:"name" binding:"required"`
	Parallelism *int32            `json:"parallelism"`
	Completions *int32            `json:"completions"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Yaml        string            `json:"yaml"`
}

type JobBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type JobPageQueryBaseDTO struct {
	PageQueryDTO
	JobBaseDTO
}

type JobApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}
