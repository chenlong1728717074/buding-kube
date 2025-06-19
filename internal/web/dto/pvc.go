package dto

type PVCCreateDTO struct {
	ClusterId        string            `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace        string            `json:"namespace" form:"namespace" binding:"required"`
	Name             string            `json:"name" form:"name" binding:"required"`
	Capacity         string            `json:"capacity" form:"capacity" binding:"required"`
	AccessModes      []string          `json:"accessModes" form:"accessModes" binding:"required"`
	StorageClassName *string           `json:"storageClassName"`
	Annotations      map[string]string `json:"annotations"`
	Labels           map[string]string `json:"labels"`
	Yaml             string            `json:"yaml"`
}

type PVCBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type PVCPageQueryBaseDTO struct {
	PageQueryDTO
	PVCBaseDTO
}

type PVCApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}
