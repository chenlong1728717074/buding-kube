package dto

type PVCreateDTO struct {
	ClusterId   string            `json:"clusterId" form:"clusterId" binding:"required"`
	Name        string            `json:"name" form:"name" binding:"required"`
	Capacity    string            `json:"capacity" form:"capacity" binding:"required"`
	AccessModes []string          `json:"accessModes" form:"accessModes" binding:"required"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Yaml        string            `json:"yaml"`
}

type PVBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type PVPageQueryBaseDTO struct {
	PageQueryDTO
	PVBaseDTO
}

type PVApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}
