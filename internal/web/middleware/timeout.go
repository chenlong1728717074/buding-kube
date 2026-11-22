package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTimeout 普通接口的超时时间. 请求上下文会传给 client-go:
// 超时或浏览器断开连接(net/http 会取消 Request.Context)时, 对集群的请求随之取消.
// skip 为不限制时间的路径前缀(终端、日志、文件传输、排空节点等长连接或耗时操作)
func RequestTimeout(timeout time.Duration, skip ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, prefix := range skip {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
