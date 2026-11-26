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

	//容器终端(WebSocket): 浏览器不能设置请求头, 认证在握手后由接口自己完成, 这里不挂 JWTAuth
	terminal := engine.Group("/terminal")
	api.NewTerminalApi(terminal)

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

	//工作负载通用接口(详情、分步创建、调整副本、版本回退等)
	workload := engine.Group("/workload", middleware.JWTAuth())
	api.NewWorkloadApi(workload)

	//service
	service := engine.Group("/service", middleware.JWTAuth())
	api.NewKubeSrvApi(service)

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

	//crd
	crd := engine.Group("/crd", middleware.JWTAuth())
	api.NewCrdApi(crd)

	//通用资源接口(列表、对象、YAML、删除), 用于 ReplicaSet / HPA / NetworkPolicy / 配额 / CRD 实例等
	resource := engine.Group("/resource", middleware.JWTAuth())
	api.NewResourceApi(resource)

	//事件中心
	event := engine.Group("/event", middleware.JWTAuth())
	api.NewEventApi(event)

	//集群功能检测(例如网络插件是否支持网络策略)
	capability := engine.Group("/capability", middleware.JWTAuth())
	api.NewCapabilityApi(capability)

	//命名空间资源拓扑图
	topology := engine.Group("/topology", middleware.JWTAuth())
	api.NewTopologyApi(topology)
}
