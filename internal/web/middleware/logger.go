package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

var middlewareLogger *log.Logger

func init() {
	middlewareLogger = log.New(os.Stdout, "[GIN] ", log.LstdFlags)
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 错误信息
		var errMsg string
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		}

		// 响应大小
		bodySize := c.Writer.Size()

		// 输出日志
		logFormat := "%s | %3d | %13v | %15s | %s | %d"

		// 根据状态码设置日志级别
		switch {
		case statusCode >= 500:
			middlewareLogger.Printf(logFormat+" | 错误信息: %s",
				reqMethod, statusCode, latencyTime, clientIP, reqUri, bodySize, errMsg)
		case statusCode >= 400:
			middlewareLogger.Printf(logFormat,
				reqMethod, statusCode, latencyTime, clientIP, reqUri, bodySize)
		case statusCode >= 300:
			middlewareLogger.Printf(logFormat,
				reqMethod, statusCode, latencyTime, clientIP, reqUri, bodySize)
		default:
			middlewareLogger.Printf(logFormat,
				reqMethod, statusCode, latencyTime, clientIP, reqUri, bodySize)
		}
	}
}
