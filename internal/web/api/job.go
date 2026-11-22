package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"

	"github.com/gin-gonic/gin"
)

type JobApi struct {
	router *gin.RouterGroup
	srv    *service.JobService
	BaseApi
}

func NewJobApi(router *gin.RouterGroup) *JobApi {
	api := JobApi{
		router: router,
		srv:    service.GetSingletonJobService(),
	}
	api.Router()
	return &api
}

func (a *JobApi) Router() {
	a.router.GET("/list", a.List)
	a.router.GET("", a.Get)
	a.router.POST("", middleware.Blocker(), BindJSON[dto.JobCreateDTO](a.Create))
	a.router.PUT("", middleware.Blocker(), BindJSON[dto.JobUpdateDTO](a.Update))
	a.router.POST("/rerun", middleware.Blocker(), BindJSON[dto.JobBaseDTO](a.Rerun))
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

func (a *JobApi) List(ctx *gin.Context) {
	var query dto.JobPageQueryBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.List(ctx.Request.Context(), query)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, BuildPageResponse(list, query.Page, query.PageSize))
}

// Get 任务详情
func (a *JobApi) Get(ctx *gin.Context) {
	var query dto.JobBaseDTO
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

// Create 创建任务
func (a *JobApi) Create(ctx *gin.Context, req dto.JobCreateDTO) {
	if err := a.srv.Create(ctx.Request.Context(), req, a.CurrentUser(ctx).Username); err != nil {
		a.InternalError(ctx, "创建失败", err)
		return
	}
	a.SuccessMsg(ctx, "创建成功")
}

// Update 编辑任务
func (a *JobApi) Update(ctx *gin.Context, req dto.JobUpdateDTO) {
	if err := a.srv.Update(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "更新失败", err)
		return
	}
	a.SuccessMsg(ctx, "更新成功")
}

// Rerun 重新执行任务
func (a *JobApi) Rerun(ctx *gin.Context, req dto.JobBaseDTO) {
	if err := a.srv.Rerun(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "重新执行失败", err)
		return
	}
	a.SuccessMsg(ctx, "已重新执行")
}

// Delete 删除任务
func (a *JobApi) Delete(ctx *gin.Context) {
	var query dto.JobBaseDTO
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
