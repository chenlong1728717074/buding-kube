package dto

// EventQueryDTO 事件中心查询; 对象条件为空时查询全部事件
type EventQueryDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	PageQueryDTO
	Namespace string `json:"namespace" form:"namespace"`
	// Type Normal / Warning
	Type string `json:"type" form:"type"`
	// Kind / Name / Uid 涉及对象的类型、名称和 UID
	Kind string `json:"kind" form:"kind"`
	Name string `json:"name" form:"name"`
	Uid  string `json:"uid" form:"uid"`
}
