package dto

type NodeQueryDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	PageQueryDTO
}

type NodeUnScheduleDTO struct {
	NodeInfoQueryDTO
	UnSchedule *bool `json:"unSchedule" form:"unSchedule" binding:"required"`
}

type NodeInfoQueryDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Hostname  string `json:"hostname" form:"hostname" binding:"required"`
}
