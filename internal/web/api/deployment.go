package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
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
	api.router.DELETE("", middleware.Blocker(), api.Delete)
	api.router.PUT("", middleware.Blocker(), api.Update)
	api.router.PUT("/rollout", middleware.Blocker(), api.Rollout)
	api.router.POST("/apply", middleware.Blocker(), api.Apply)
}

func (api *DeploymentApi) List(ctx *gin.Context) {
	var query dto.WorkloadQueryDTO
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

func (api *DeploymentApi) Update(ctx *gin.Context) {
	var up dto.WorkloadUpdateDTO
	if err := ctx.ShouldBindJSON(&up); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Update(up)
	if err != nil {
		api.InternalError(ctx, "修改失败:", err)
		return
	}
	api.SuccessMsg(ctx, "修改成功")

}

func (api *DeploymentApi) Delete(ctx *gin.Context) {
	var base dto.WorkloadBaseDTO
	if err := ctx.ShouldBindJSON(&base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Delete(base)
	if err != nil {
		api.InternalError(ctx, "删除失败:", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}

func (api *DeploymentApi) Rollout(ctx *gin.Context) {
	var rollout dto.WorkloadBaseDTO
	if err := ctx.ShouldBindJSON(&rollout); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Rollout(rollout)
	if err != nil {
		api.InternalError(ctx, "操作失败:", err)
		return
	}
	api.SuccessMsg(ctx, "操作成功")

}

func (api *DeploymentApi) Apply(ctx *gin.Context) {
	var apply dto.WorkloadApplyDTO
	if err := ctx.ShouldBindJSON(&apply); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Apply(apply)
	if err != nil {
		api.InternalError(ctx, "操作失败:", err)
		return
	}
	api.SuccessMsg(ctx, "操作成功")
}
