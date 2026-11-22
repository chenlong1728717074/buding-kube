package dto

// CrdPageQueryDTO CRD 分页查询参数
type CrdPageQueryDTO struct {
	PageQueryDTO
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	// Refresh 忽略缓存重新查询(点击刷新、修改 CRD 之后)
	Refresh bool `json:"refresh" form:"refresh"`
}

// CrdBaseDTO CRD 基础参数(详情/删除)
type CrdBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Name      string `json:"name" form:"name" binding:"required"`
}
