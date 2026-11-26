package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"

	"github.com/gin-gonic/gin"
)

type TopologyApi struct {
	router *gin.RouterGroup
	srv    *service.TopologyService
	BaseApi
}

func NewTopologyApi(router *gin.RouterGroup) *TopologyApi {
	api := TopologyApi{
		router: router,
		srv:    service.GetSingletonTopologyService(),
	}
	api.Router()
	return &api
}

func (api *TopologyApi) Router() {
	api.router.GET("", api.GetNamespaceTopology)
}

func (api *TopologyApi) GetNamespaceTopology(ctx *gin.Context) {
	var query dto.TopologyQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.GetNamespaceTopology(ctx.Request.Context(), query)
	if err != nil {
		api.InternalError(ctx, "获取资源拓扑图失败:", err)
		return
	}
	api.SuccessWithData(ctx, result)
}
