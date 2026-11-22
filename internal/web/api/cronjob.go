package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"

	"github.com/gin-gonic/gin"
)

type CronJobApi struct {
	router *gin.RouterGroup
	srv    *service.CronJobService
	BaseApi
}

func NewCronJobApi(router *gin.RouterGroup) *CronJobApi {
	api := CronJobApi{
		router: router,
		srv:    service.GetSingletonCronJobService(),
	}
	api.Router()
	return &api
}

func (a *CronJobApi) Router() {
	a.router.GET("/list", a.List)
	a.router.GET("", a.Get)
	a.router.POST("", middleware.Blocker(), BindJSON[dto.CronJobCreateDTO](a.Create))
	a.router.PUT("", middleware.Blocker(), BindJSON[dto.CronJobUpdateDTO](a.Update))
	a.router.PUT("/suspend", middleware.Blocker(), BindJSON[dto.CronJobSuspendDTO](a.Suspend))
	a.router.POST("/trigger", middleware.Blocker(), BindJSON[dto.CronJobBaseDTO](a.Trigger))
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

func (a *CronJobApi) List(ctx *gin.Context) {
	var query dto.CronJobPageQueryBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.List(query)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, BuildPageResponse(list, query.Page, query.PageSize))
}

// Get 定时任务详情
func (a *CronJobApi) Get(ctx *gin.Context) {
	var query dto.CronJobBaseDTO
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

// Create 创建定时任务
func (a *CronJobApi) Create(ctx *gin.Context, req dto.CronJobCreateDTO) {
	if err := a.srv.Create(ctx.Request.Context(), req, a.CurrentUser(ctx).Username); err != nil {
		a.InternalError(ctx, "创建失败", err)
		return
	}
	a.SuccessMsg(ctx, "创建成功")
}

// Update 编辑定时任务
func (a *CronJobApi) Update(ctx *gin.Context, req dto.CronJobUpdateDTO) {
	if err := a.srv.Update(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "更新失败", err)
		return
	}
	a.SuccessMsg(ctx, "更新成功")
}

// Suspend 暂停 / 启用定时任务
func (a *CronJobApi) Suspend(ctx *gin.Context, req dto.CronJobSuspendDTO) {
	if err := a.srv.Suspend(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "操作失败", err)
		return
	}
	if req.Suspend {
		a.SuccessMsg(ctx, "已暂停")
	} else {
		a.SuccessMsg(ctx, "已启用")
	}
}

// Trigger 立即执行一次
func (a *CronJobApi) Trigger(ctx *gin.Context, req dto.CronJobBaseDTO) {
	jobName, err := a.srv.Trigger(ctx.Request.Context(), req, a.CurrentUser(ctx).Username)
	if err != nil {
		a.InternalError(ctx, "执行失败", err)
		return
	}
	a.SuccessWithData(ctx, gin.H{"jobName": jobName})
}

// Delete 删除定时任务
func (a *CronJobApi) Delete(ctx *gin.Context) {
	var query dto.CronJobBaseDTO
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
