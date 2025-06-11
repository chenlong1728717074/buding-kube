package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// 统一响应码
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

type BaseApi struct {
}

// BindJSON 绑定JSON请求
func (api *BaseApi) BindJSON(c *gin.Context, obj interface{}) error {
	return c.ShouldBindJSON(obj)
}

// BindForm 绑定表单请求
func (api *BaseApi) BindForm(c *gin.Context, obj interface{}) error {
	return c.ShouldBind(obj)
}

// GetParam 获取RESTful风格的URL参数
func (api *BaseApi) GetParam(c *gin.Context, key string) string {
	return c.Param(key)
}

func (api *BaseApi) GetParamUint(c *gin.Context, key string) (uint, error) {
	param := c.Param(key)
	val, err := strconv.ParseUint(param, 10, 64)
	return uint(val), err
}

func (api *BaseApi) GetParamWithConvert(c *gin.Context, key string, converter func(string) (interface{}, error)) (interface{}, error) {
	param := c.Param(key)
	return converter(param)
}

// GetQuery 获取查询参数
func (api *BaseApi) GetQuery(c *gin.Context, key string) (string, bool) {
	return c.GetQuery(key)
}

func (api *BaseApi) GetQueryUint(c *gin.Context, key string) (uint, error) {
	param, flag := c.GetQuery(key)
	if !flag {
		return 0, errors.New("获取不到id参数")
	}
	val, err := strconv.ParseUint(param, 10, 64)
	return uint(val), err
}

// Success 成功响应
func (api *BaseApi) Success(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	})
}

// SuccessWithData 仅数据的成功响应
func (api *BaseApi) SuccessWithData(c *gin.Context, data interface{}) {
	api.Success(c, "操作成功", data)
}

// SuccessMsg 仅消息的成功响应
func (api *BaseApi) SuccessMsg(c *gin.Context, msg string) {
	api.Success(c, msg, nil)
}

// Fail 失败响应
func (api *BaseApi) Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// FailWithError 带错误信息的失败响应
func (api *BaseApi) FailWithError(c *gin.Context, code int, msg string, err error) {
	if err != nil {
		msg = msg + ": " + err.Error()
	}
	api.Fail(c, code, msg)
}

// ParamBindError 参数绑定错误
func (api *BaseApi) ParamBindError(c *gin.Context, err error) {
	api.FailWithError(c, CodeInvalidParams, "参数绑定失败", err)
}

// InternalError 内部服务器错误
func (api *BaseApi) InternalError(c *gin.Context, msg string, err error) {
	api.FailWithError(c, CodeInternalError, msg, err)
}

// NotFound 资源不存在错误
func (api *BaseApi) NotFound(c *gin.Context, msg string) {
	api.Fail(c, CodeNotFound, msg)
}

// Unauthorized 未授权错误
func (api *BaseApi) Unauthorized(c *gin.Context, msg string) {
	api.Fail(c, CodeUnauthorized, msg)
}

// Forbidden 权限不足错误
func (api *BaseApi) Forbidden(c *gin.Context, msg string) {
	api.Fail(c, CodeForbidden, msg)
}

// ParamError 参数错误
func (api *BaseApi) ParamError(c *gin.Context, msg string) {
	api.Fail(c, CodeInvalidParams, msg)
}
