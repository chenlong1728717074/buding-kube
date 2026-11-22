package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

type KubeSrvApi struct {
	router *gin.RouterGroup
	srv    *service.KubeSrvService
	BaseApi
}

func NewKubeSrvApi(router *gin.RouterGroup) *KubeSrvApi {
	api := KubeSrvApi{
		router: router,
		srv:    service.GetSingletonKubeSrvService(),
	}
	api.Router()
	return &api
}

func (api *KubeSrvApi) Router() {
	api.router.GET("/list", api.List)
	api.router.GET("", api.GetInfo)
	api.router.POST("", middleware.Blocker(), BindJSON[dto.ServiceFormDTO](api.Create))
	api.router.PUT("", middleware.Blocker(), BindJSON[dto.ServiceFormDTO](api.Update))
	api.router.DELETE("", middleware.Blocker(), api.Delete)
}

// Delete 删除Service
func (api *KubeSrvApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	if err := api.srv.Delete(ctx.Request.Context(), base); err != nil {
		api.InternalError(ctx, "删除失败:%v", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}

func (api *KubeSrvApi) List(ctx *gin.Context) {
	var query dto.ServiceQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	list, err := api.srv.List(ctx.Request.Context(), query)
	if err != nil {
		api.InternalError(ctx, "查询失败:", err)
		return
	}
	response := BuildPageResponse(list, query.Page, query.PageSize)
	api.SuccessWithData(ctx, response)
}

func (api *KubeSrvApi) GetInfo(ctx *gin.Context) {
	var query dto.ServiceBaseDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	info, err := api.srv.Get(ctx.Request.Context(), query)
	if err != nil {
		api.InternalError(ctx, "获取详情失败:", err)
		return
	}
	api.SuccessWithData(ctx, info)
}

// Create 创建服务
func (api *KubeSrvApi) Create(ctx *gin.Context, req dto.ServiceFormDTO) {
	if err := api.srv.Create(ctx.Request.Context(), req, api.CurrentUser(ctx).Username); err != nil {
		api.InternalError(ctx, "创建失败", err)
		return
	}
	api.SuccessMsg(ctx, "创建成功")
}

// Update 编辑服务
func (api *KubeSrvApi) Update(ctx *gin.Context, req dto.ServiceFormDTO) {
	if err := api.srv.Update(ctx.Request.Context(), req); err != nil {
		api.InternalError(ctx, "更新失败", err)
		return
	}
	api.SuccessMsg(ctx, "更新成功")
}
