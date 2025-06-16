package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS,HEAD,PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
	//return cors.New(cors.Config{
	//	AllowOrigins:     []string{"*"},
	//	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	//	AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	//	ExposeHeaders:    []string{"Content-Length"},
	//	AllowCredentials: false,
	//	MaxAge:           12 * time.Hour,
	//})
}
