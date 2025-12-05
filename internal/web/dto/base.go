package dto

// PageQueryDTO 分页查询基础参数
type PageQueryDTO struct {
	Page     int    `form:"page" json:"page" example:"1"`          // 页码，默认为1
	PageSize int    `form:"pageSize" json:"pageSize" example:"10"` // 每页数量，默认为10
	Keyword  string `form:"keyword" json:"keyword" example:"关键词"`  // 搜索关键词
}

type BaseYamlApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Yaml      string `json:"yaml" binding:"required"`
}

type BaseDTO struct {
	// 集群ID
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	// Name 资源名称
	Name string `json:"name"`
	// Namespace 命名空间
	Namespace string `json:"namespace"`
}

type BaseInfoDTO struct {
	BaseDTO
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
}

type ResourcePageQueryBaseDTO struct {
	PageQueryDTO
	BaseDTO
}
