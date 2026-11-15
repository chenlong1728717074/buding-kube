package dto

// CrdPageQueryDTO CRD 分页查询参数
type CrdPageQueryDTO struct {
	PageQueryDTO
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
}

// CrdBaseDTO CRD 基础参数(详情/删除)
type CrdBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Name      string `json:"name" form:"name" binding:"required"`
}
