package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"

	"github.com/gin-gonic/gin"
)

// WorkloadApi Deployment / StatefulSet / DaemonSet 通用接口, 用 kind 区分类型
type WorkloadApi struct {
	router *gin.RouterGroup
	srv    *service.WorkloadService
	BaseApi
}

func NewWorkloadApi(router *gin.RouterGroup) *WorkloadApi {
	api := WorkloadApi{
		router: router,
		srv:    service.GetSingletonWorkloadService(),
	}
	api.Router()
	return &api
}

func (a *WorkloadApi) Router() {
	a.router.GET("", a.Get)
	a.router.GET("/selectors", a.Selectors)
	a.router.POST("", middleware.Blocker(), BindJSON[dto.WorkloadCreateDTO](a.Create))
	a.router.PUT("/info", middleware.Blocker(), BindJSON[dto.WorkloadInfoDTO](a.UpdateInfo))
	a.router.PUT("/scale", middleware.Blocker(), BindJSON[dto.WorkloadScaleDTO](a.Scale))
	a.router.PUT("/image", middleware.Blocker(), BindJSON[dto.WorkloadImageDTO](a.UpdateImages))
	a.router.PUT("/restart", middleware.Blocker(), BindJSON[dto.WorkloadTargetDTO](a.Restart))
	a.router.PUT("/rollback", middleware.Blocker(), BindJSON[dto.WorkloadRollbackDTO](a.Rollback))
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

// Get 工作负载详情
func (a *WorkloadApi) Get(ctx *gin.Context) {
	var query dto.WorkloadTargetDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	detail, err := a.srv.Get(ctx.Request.Context(), query)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, detail)
}

// Selectors 命名空间中的工作负载及其容器组标签
func (a *WorkloadApi) Selectors(ctx *gin.Context) {
	var query dto.NamespaceBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.Selectors(ctx.Request.Context(), query.ClusterId, query.Namespace)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, list)
}

// Create 创建工作负载
func (a *WorkloadApi) Create(ctx *gin.Context, req dto.WorkloadCreateDTO) {
	if err := a.srv.Create(ctx.Request.Context(), req, a.CurrentUser(ctx).Username); err != nil {
		a.InternalError(ctx, "创建失败", err)
		return
	}
	a.SuccessMsg(ctx, "创建成功")
}

func (a *WorkloadApi) UpdateInfo(ctx *gin.Context, req dto.WorkloadInfoDTO) {
	if err := a.srv.UpdateInfo(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "修改失败", err)
		return
	}
	a.SuccessMsg(ctx, "修改成功")
}

func (a *WorkloadApi) Scale(ctx *gin.Context, req dto.WorkloadScaleDTO) {
	if err := a.srv.Scale(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "调整副本数失败", err)
		return
	}
	a.SuccessMsg(ctx, "副本数已调整")
}

func (a *WorkloadApi) UpdateImages(ctx *gin.Context, req dto.WorkloadImageDTO) {
	if err := a.srv.UpdateImages(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "更新镜像失败", err)
		return
	}
	a.SuccessMsg(ctx, "镜像已更新")
}

func (a *WorkloadApi) Restart(ctx *gin.Context, req dto.WorkloadTargetDTO) {
	if err := a.srv.Restart(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "重新创建失败", err)
		return
	}
	a.SuccessMsg(ctx, "已开始重新创建容器组")
}

func (a *WorkloadApi) Rollback(ctx *gin.Context, req dto.WorkloadRollbackDTO) {
	if err := a.srv.Rollback(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "版本回退失败", err)
		return
	}
	a.SuccessMsg(ctx, "已回退")
}

func (a *WorkloadApi) Delete(ctx *gin.Context) {
	var query dto.WorkloadTargetDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	if err := a.srv.Delete(ctx.Request.Context(), query); err != nil {
		a.InternalError(ctx, "删除失败", err)
		return
	}
	a.SuccessMsg(ctx, "删除成功")
}
