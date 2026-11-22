package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"

	"github.com/gin-gonic/gin"
)

type StorageClassApi struct {
	router *gin.RouterGroup
	srv    *service.StorageClassService
	BaseApi
}

func NewStorageClassApi(router *gin.RouterGroup) *StorageClassApi {
	api := StorageClassApi{
		router: router,
		srv:    service.GetSingletonStorageClassService(),
	}
	api.Router()
	return &api
}

func (a *StorageClassApi) Router() {
	a.router.GET("/list", a.List)
	a.router.GET("", a.Get)
	a.router.POST("", middleware.Blocker(), BindJSON[dto.StorageClassCreateDTO](a.Create))
	a.router.PUT("", middleware.Blocker(), BindJSON[dto.StorageClassUpdateDTO](a.Update))
	a.router.POST("/apply", middleware.Blocker(), BindJSON[dto.StorageClassApplyDTO](a.Apply))
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

// Get 存储类详情(含使用该存储类的持久卷)
func (a *StorageClassApi) Get(ctx *gin.Context) {
	var query dto.StorageClassBaseDTO
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

// Create 创建存储类
func (a *StorageClassApi) Create(ctx *gin.Context, req dto.StorageClassCreateDTO) {
	if err := a.srv.Create(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "创建失败", err)
		return
	}
	a.SuccessMsg(ctx, "创建成功")
}

// Update 编辑存储类
func (a *StorageClassApi) Update(ctx *gin.Context, req dto.StorageClassUpdateDTO) {
	if err := a.srv.Update(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "更新失败", err)
		return
	}
	a.SuccessMsg(ctx, "更新成功")
}

// Apply 通过 YAML 创建或更新存储类
func (a *StorageClassApi) Apply(ctx *gin.Context, req dto.StorageClassApplyDTO) {
	if err := a.srv.Apply(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "执行失败", err)
		return
	}
	a.SuccessMsg(ctx, "执行成功")
}

// Delete 删除StorageClass
func (a *StorageClassApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := a.BindQuery(ctx, &base); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	if err := a.srv.Delete(ctx.Request.Context(), base); err != nil {
		a.InternalError(ctx, "删除失败", err)
		return
	}
	a.SuccessMsg(ctx, "删除成功")
}

func (a *StorageClassApi) List(ctx *gin.Context) {
	var query dto.StorageClassPageQueryBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.List(ctx.Request.Context(), query)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	response := BuildPageResponse(list, query.Page, query.PageSize)
	a.SuccessWithData(ctx, response)
}
