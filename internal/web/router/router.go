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

	//用户管理
	user := engine.Group("/user", middleware.JWTAuth())
	api.NewUserApi(user)

	//deployment
	deployment := engine.Group("/deployment", middleware.JWTAuth())
	api.NewDeploymentApi(deployment)

	//statefulSet
	statefulSet := engine.Group("/statefulSet", middleware.JWTAuth())
	api.NewStatefulSetApi(statefulSet)

	//daemonSet
	daemonSet := engine.Group("/daemonSet", middleware.JWTAuth())
	api.NewDaemonSetApi(daemonSet)

	//service
	service := engine.Group("/service", middleware.JWTAuth())
	api.NewKubeSrvApi(service)

	//replicaSet
	replicaSet := engine.Group("/replicaSet", middleware.JWTAuth())
	api.NewReplicaSetApi(replicaSet)

	//configMap
	configMap := engine.Group("/configMap", middleware.JWTAuth())
	api.NewConfigMapApi(configMap)

	//secret
	secret := engine.Group("/secret", middleware.JWTAuth())
	api.NewSecretApi(secret)

	//pv
	pv := engine.Group("/pv", middleware.JWTAuth())
	api.NewPVApi(pv)

	//pvc
	pvc := engine.Group("/pvc", middleware.JWTAuth())
	api.NewPVCApi(pvc)

	//storageClass
	storageClass := engine.Group("/storageClass", middleware.JWTAuth())
	api.NewStorageClassApi(storageClass)

	//serviceAccount
	serviceAccount := engine.Group("/serviceAccount", middleware.JWTAuth())
	api.NewServiceAccountApi(serviceAccount)

	//ingress
	ingress := engine.Group("/ingress", middleware.JWTAuth())
	api.NewIngressApi(ingress)

	//endpoint (已过时，建议使用EndpointSlices)
	endpoint := engine.Group("/endpoint", middleware.JWTAuth())
	api.NewEndpointApi(endpoint)

	//endpointSlice
	endpointSlice := engine.Group("/endpointslice", middleware.JWTAuth())
	api.NewEndpointSliceApi(endpointSlice)

	//cronJob
	cronJob := engine.Group("/cronJob", middleware.JWTAuth())
	api.NewCronJobApi(cronJob)

	//job
	job := engine.Group("/job", middleware.JWTAuth())
	api.NewJobApi(job)

	//apply
	apply := engine.Group("/apply", middleware.JWTAuth())
	api.NewApplyApi(apply)

	//tool
	tool := engine.Group("/tool", middleware.JWTAuth())
	api.NewToolApi(tool)
}
