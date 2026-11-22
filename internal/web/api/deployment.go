package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"

	"github.com/gin-gonic/gin"
)

// DeploymentApi Deployment 列表; 其他操作统一使用 /workload 接口
type DeploymentApi struct {
	router *gin.RouterGroup
	srv    *service.DeploymentService
	BaseApi
}

func NewDeploymentApi(router *gin.RouterGroup) *DeploymentApi {
	api := DeploymentApi{
		router: router,
		srv:    service.GetSingletonDeploymentService(),
	}
	api.Router()
	return &api
}

func (api *DeploymentApi) Router() {
	api.router.GET("/list", api.List)
}

func (api *DeploymentApi) List(ctx *gin.Context) {
	var query dto.WorkloadQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	list, err := api.srv.List(ctx.Request.Context(), query)
	if err != nil {
		api.InternalError(ctx, "查询失败", err)
		return
	}
	api.SuccessWithData(ctx, BuildPageResponse(list, query.Page, query.PageSize))
}
