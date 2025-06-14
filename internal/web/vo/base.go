package vo

import (
	"net/http"
)

const (
	CodeSuccess            = http.StatusOK
	CodeInvalidParams      = http.StatusBadRequest
	CodeUnauthorized       = http.StatusUnauthorized
	CodeForbidden          = http.StatusForbidden
	CodeNotFound           = http.StatusNotFound
	CodeInternalError      = http.StatusInternalServerError
	CodeServiceUnavailable = http.StatusServiceUnavailable
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code" example:"200"`    // 响应码
	Msg  string      `json:"msg" example:"success"` // 响应消息
	Data interface{} `json:"data"`                  // 响应数据
}

// PageResponse 分页响应结构
type PageResponse struct {
	Items     interface{} `json:"items"`                  // 数据项
	Total     int64       `json:"total" example:"100"`    // 总数
	Page      int         `json:"page" example:"1"`       // 当前页码
	PageSize  int         `json:"pageSize" example:"10"`  // 每页大小
	TotalPage int         `json:"totalPage" example:"10"` // 总页数
}
