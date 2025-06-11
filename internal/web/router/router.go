package router

import (
	"buding-kube/internal/web/api"
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置所有路由
func SetupRouter(engine *gin.RouterGroup) {
	//集群管理
	cluster := engine.Group("/cluster")
	api.NewClusterApi(cluster)
	//命名空间
	namespace := engine.Group("/namespace")
	api.NewNamespacesApi(namespace)
}
