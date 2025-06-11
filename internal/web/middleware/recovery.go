package middleware

import (
	"buding-kube/internal/web/vo"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime/debug"
)

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 打印堆栈信息
				stackInfo := string(debug.Stack())
				middlewareLogger.Printf("处理请求发生panic: %v\n堆栈信息:\n%s", err, stackInfo)

				// 判断错误类型
				var errMsg string
				switch err := err.(type) {
				case error:
					errMsg = err.Error()
				case string:
					errMsg = err
				default:
					errMsg = fmt.Sprintf("%v", err)
				}

				// 返回500错误响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, vo.Response{
					Code: vo.CodeInternalError,
					Msg:  "服务器内部错误: " + errMsg,
					Data: nil,
				})
			}
		}()

		// 继续处理请求
		c.Next()
	}
}
