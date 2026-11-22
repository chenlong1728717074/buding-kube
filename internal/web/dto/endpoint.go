package dto

// EndpointBaseDTO 已过时，建议使用 EndpointSliceBaseDTO
// Deprecated: Use EndpointSliceBaseDTO instead
type EndpointBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

// EndpointPageQueryBaseDTO 已过时，建议使用 EndpointSlicePageQueryBaseDTO
// Deprecated: Use EndpointSlicePageQueryBaseDTO instead
type EndpointPageQueryBaseDTO struct {
	PageQueryDTO
	EndpointBaseDTO
}
