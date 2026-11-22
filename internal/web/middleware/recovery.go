package middleware

import (
	"buding-kube/internal/web/vo"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// http.ErrAbortHandler 表示主动中断连接(如文件流传输到一半失败), 交还给 net/http 静默处理
			if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(rec)
			}

			// 打印堆栈信息
			middlewareLogger.Printf("处理请求发生panic: %v\n堆栈信息:\n%s", rec, debug.Stack())

			// 响应已经开始写出(流式下载/日志), 不能再追加 JSON, 否则会污染响应内容
			if c.Writer.Written() {
				c.Abort()
				return
			}

			var errMsg string
			switch e := rec.(type) {
			case error:
				errMsg = e.Error()
			case string:
				errMsg = e
			default:
				errMsg = fmt.Sprintf("%v", e)
			}

			// 返回500错误响应
			c.AbortWithStatusJSON(http.StatusInternalServerError, vo.Response{
				Code: vo.CodeInternalError,
				Msg:  "服务器内部错误: " + errMsg,
				Data: nil,
			})
		}()

		// 继续处理请求
		c.Next()
	}
}
