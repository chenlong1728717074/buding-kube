package api

import (
	"buding-kube/internal/service"

	"github.com/gin-gonic/gin"
)

// CapabilityApi 检测集群对某些功能的支持情况
type CapabilityApi struct {
	router *gin.RouterGroup
	BaseApi
}

func NewCapabilityApi(router *gin.RouterGroup) *CapabilityApi {
	api := CapabilityApi{router: router}
	api.Router()
	return &api
}

func (a *CapabilityApi) Router() {
	a.router.GET("/networkPolicy", a.NetworkPolicy)
}

type clusterQuery struct {
	ClusterId string `form:"clusterId" binding:"required"`
}

// NetworkPolicy 网络插件是否支持网络策略
func (a *CapabilityApi) NetworkPolicy(ctx *gin.Context) {
	var query clusterQuery
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	result, err := service.NetworkPolicySupport(ctx.Request.Context(), query.ClusterId)
	if err != nil {
		a.InternalError(ctx, "检测失败", err)
		return
	}
	a.SuccessWithData(ctx, result)
}
