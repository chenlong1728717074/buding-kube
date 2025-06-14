package dto

// PageQueryDTO 分页查询基础参数
type PageQueryDTO struct {
	Page     int    `form:"page" json:"page" example:"1"`          // 页码，默认为1
	PageSize int    `form:"pageSize" json:"pageSize" example:"10"` // 每页数量，默认为10
	Keyword  string `form:"keyword" json:"keyword" example:"关键词"`  // 搜索关键词
}
