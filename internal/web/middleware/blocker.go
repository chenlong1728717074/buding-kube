package middleware

import (
	"buding-kube/internal/web/vo"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

// Blocker 用于演示环境阻断敏感操作
func Blocker() gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("APP_ENV") == "demo" {
			c.AbortWithStatusJSON(http.StatusOK, vo.Response{
				Code: vo.CodeInternalError,
				Msg:  "演示环境阻断:不允许操作该接口 ",
				Data: nil,
			})
			c.Abort()
		}
		c.Next()
	}
}
