package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func Cors() gin.HandlerFunc {
	//return func(c *gin.Context) {
	//	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	//	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	//	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	//
	//	// 处理预检请求
	//	if c.Request.Method == "OPTIONS" {
	//		c.AbortWithStatus(http.StatusNoContent)
	//		return
	//	}
	//	c.Next()
	//}
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
			"X-Auth-Token",
			"Cache-Control",
			"Connection",
			"Cookie",
			"Host",
			"Pragma",
			"Referer",
			"User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Cache-Control",
			"Content-Language",
			"Content-Disposition",
			"Access-Control-Allow-Origin",
			"Authorization",
		},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}
