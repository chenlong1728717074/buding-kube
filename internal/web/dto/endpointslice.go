package dto

type EndpointSliceBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type EndpointSlicePageQueryBaseDTO struct {
	PageQueryDTO
	EndpointSliceBaseDTO
}
