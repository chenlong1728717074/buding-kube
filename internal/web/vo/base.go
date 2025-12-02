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

	ErrorEvent = "error"
	StartEvent = "start"
	PingEvent  = "ping"
	ChunkEvent = "chunk"
	DoneEvent  = "done"
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

type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type SSEChan chan SSEEvent

func (c SSEChan) Close() {
	close(c)
}
func (c SSEChan) SSEStart(msg string, data interface{}) {
	c <- SSEEvent{
		Event: StartEvent,
		Data: Response{
			Code: CodeSuccess,
			Msg:  msg,
			Data: data,
		},
	}
}
func (c SSEChan) SSEHeartbeat() {
	c <- SSEEvent{
		Event: PingEvent,
		Data: Response{
			Code: CodeSuccess,
			Msg:  "ping",
			Data: nil,
		},
	}
}

func (c SSEChan) SSEError(err error) {
	c <- SSEEvent{
		Event: ErrorEvent,
		Data: Response{
			Code: CodeInternalError,
			Msg:  "获取日志流错误:" + err.Error(),
			Data: nil,
		},
	}
}
func (c SSEChan) SSEDone() {
	c <- SSEEvent{
		Event: DoneEvent,
		Data: Response{
			Code: CodeSuccess,
			Msg:  "结束连接",
			Data: nil,
		},
	}
}

func (c SSEChan) SSEChunk(chunk string) {
	c <- SSEEvent{
		Event: ChunkEvent,
		Data: Response{
			Code: CodeSuccess,
			Msg:  "事件消息",
			Data: chunk,
		},
	}
}

type BaseVO struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
