package middleware

import (
	"buding-kube/internal/web/vo"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

// Blocker 用于演示环境阻断敏感操作.
// 环境变量在注册路由时读取一次即可, 不必每个请求都读取
func Blocker() gin.HandlerFunc {
	isDemo := os.Getenv("APP_ENV") == "demo"
	return func(c *gin.Context) {
		if isDemo {
			c.AbortWithStatusJSON(http.StatusOK, vo.Response{
				Code: vo.CodeInternalError,
				Msg:  "演示环境阻断:不允许操作该接口",
				Data: nil,
			})
			return
		}
		c.Next()
	}
}
