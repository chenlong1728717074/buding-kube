package dto

// EndpointCreateDTO 已过时，建议使用 EndpointSliceCreateDTO
// Deprecated: Use EndpointSliceCreateDTO instead
type EndpointCreateDTO struct {
	ClusterId   string            `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace   string            `json:"namespace" form:"namespace" binding:"required"`
	Name        string            `json:"name" form:"name" binding:"required"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Yaml        string            `json:"yaml"`
}

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

// EndpointApplyDTO 已过时，建议使用 EndpointSliceApplyDTO
// Deprecated: Use EndpointSliceApplyDTO instead
type EndpointApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}
