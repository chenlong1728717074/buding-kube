package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

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
	var query dto.DeploymentQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	list, err := api.srv.List(query)
	if err != nil {
		api.InternalError(ctx, "查询失败:", err)
		return
	}
	response := BuildPageResponse(list, query.Page, query.PageSize)
	api.SuccessWithData(ctx, response)
}
