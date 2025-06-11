package router

import (
	"buding-kube/internal/web/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

// SetupRouter 配置所有路由
func SetupRouter(engine *gin.RouterGroup) {
	// 配置CORS中间件
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//集群管理
	cluster := engine.Group("/cluster")
	api.NewClusterApi(cluster)
}
