package dto

import "net/http"

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
	Code int         `json:"code"` // 响应码
	Msg  string      `json:"msg"`  // 响应消息
	Data interface{} `json:"data"` // 响应数据
}
