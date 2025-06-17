package api

import (
	"buding-kube/internal/model"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils/jwt"
	"bufio"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"time"
)

// 统一响应码

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

// BindQuery 绑定Query请求
func (api *BaseApi) BindQuery(c *gin.Context, obj interface{}) error {
	return c.ShouldBindQuery(obj)
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
	c.JSON(http.StatusOK, vo.Response{
		Code: vo.CodeSuccess,
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
	c.JSON(http.StatusOK, vo.Response{
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
	api.FailWithError(c, vo.CodeInvalidParams, "参数绑定失败", err)
}

// InternalError 内部服务器错误
func (api *BaseApi) InternalError(c *gin.Context, msg string, err error) {
	api.FailWithError(c, vo.CodeInternalError, msg, err)
}

// NotFound 资源不存在错误
func (api *BaseApi) NotFound(c *gin.Context, msg string) {
	api.Fail(c, vo.CodeNotFound, msg)
}

// Unauthorized 未授权错误
func (api *BaseApi) Unauthorized(c *gin.Context, msg string) {
	api.Fail(c, vo.CodeUnauthorized, msg)
}

// Forbidden 权限不足错误
func (api *BaseApi) Forbidden(c *gin.Context, msg string) {
	api.Fail(c, vo.CodeForbidden, msg)
}

// ParamError 参数错误
func (api *BaseApi) ParamError(c *gin.Context, msg string) {
	api.Fail(c, vo.CodeInvalidParams, msg)
}

func BuildPageResponse[T any](data []T, page, pageSize int) vo.PageResponse {
	total := len(data)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算起止位置
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	// 切片
	pagedItems := data[start:end]

	totalPage := (total + pageSize - 1) / pageSize // 向上取整

	return vo.PageResponse{
		Items:     pagedItems,
		Total:     int64(total),
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}
}

func (api *BaseApi) CurrentUser(ctx *gin.Context) (*model.User, error) {
	claims, exists := ctx.Get("claims")
	if !exists {
		api.Unauthorized(ctx, "未认证")
		return nil, errors.New("未认证")
	}

	// 将声明转换为JWT声明对象
	jwtClaims, ok := claims.(*jwt.Claims)
	if !ok {
		api.InternalError(ctx, "无效的JWT声明", nil)
		return nil, errors.New("无效的JWT声明")
	}
	return &model.User{
		Username: jwtClaims.Username,
		Role:     jwtClaims.Role,
		Cluster:  jwtClaims.Cluster,
	}, nil
}

func (api *BaseApi) keepAliveSSE(ctx *gin.Context) {
	_ = ctx.Request.ParseMultipartForm(32 << 20)
	// 设置SSE头信息
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache,private")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.Header(`Access-Control-Allow-Origin`, `*`)
}

func (api *BaseApi) keepAliveTextPlain(ctx *gin.Context) {
	_ = ctx.Request.ParseMultipartForm(32 << 20)
	// 设置SSE头信息
	ctx.Header("Content-Type", "text/plain; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache,private")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no") // 禁用 Nginx 的缓冲
	ctx.Header(`Access-Control-Allow-Origin`, `*`)
	ctx.Header("Transfer-Encoding", "chunked")
	// 设置状态码
	ctx.Writer.WriteHeader(200)
}

type TimeoutReader struct {
	read    *bufio.Reader
	timeout time.Duration
}

// TimeoutError 是一个自定义错误类型，实现了 net.Error 接口
type TimeoutError struct {
	err error
}

// 创建一个新的超时错误
func NewTimeoutError(msg string) *TimeoutError {
	return &TimeoutError{
		err: errors.New(msg),
	}
}

// Error 实现 error 接口
func (e *TimeoutError) Error() string {
	return e.err.Error()
}

// Timeout 实现 net.Error 接口，始终返回 true 表示这是一个超时错误
func (e *TimeoutError) Timeout() bool {
	return true
}

// Temporary 实现 net.Error 接口，虽然已废弃但仍需实现
// 对于超时错误，通常也认为是临时性的
func (e *TimeoutError) Temporary() bool {
	return true
}

// ErrReadTimeout 预定义的超时错误实例
var ErrReadTimeout = NewTimeoutError("read timeout")

func NewTimeoutReader(r io.Reader, timeout time.Duration) *TimeoutReader {
	return &TimeoutReader{read: bufio.NewReader(r), timeout: timeout}
}

func (tr *TimeoutReader) Read() (msg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), tr.timeout)
	defer cancel()

	lineChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		line, err := tr.read.ReadString('\n')
		if err != nil {
			errorChan <- err
			return
		}
		lineChan <- line
	}()

	select {
	case line := <-lineChan:
		return line, nil
	case err := <-errorChan:
		return "", err
	case <-ctx.Done():
		// 读取超时
		return "", ErrReadTimeout
	}
}

type StreamPlan struct {
	writer gin.ResponseWriter
}

func NewStreamPlan(ctx *gin.Context) *StreamPlan {
	return &StreamPlan{
		writer: ctx.Writer,
	}
}

func (sp StreamPlan) Writer(msg string) {
	_, err := sp.writer.Write([]byte(msg))
	if err != nil {
		logs.Error("连接断开 %s ", err.Error())
		return
	}
	sp.writer.Flush()
	return
}
