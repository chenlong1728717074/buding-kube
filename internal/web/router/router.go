package router

import (
	"buding-kube/internal/web/api"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置所有路由
func SetupRouter(engine *gin.RouterGroup) {
	//集群管理
	cluster := engine.Group("/cluster", middleware.JWTAuth())
	api.NewClusterApi(cluster)
	//命名空间
	namespace := engine.Group("/namespace", middleware.JWTAuth())
	api.NewNamespacesApi(namespace)
	//节点
	node := engine.Group("/node", middleware.JWTAuth())
	api.NewNodeApi(node)
	//pod
	pod := engine.Group("/pod", middleware.JWTAuth())
	api.NewPodApi(pod)
	//auth
	auth := engine.Group("/auth")
	api.NewAuthApi(auth)
}
