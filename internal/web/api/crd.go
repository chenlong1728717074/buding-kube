package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"buding-kube/pkg/logs"

	"github.com/gin-gonic/gin"
)

// CrdApi CRD 定义; 自定义资源(CRD 的实例)使用通用的 /resource 接口
type CrdApi struct {
	router *gin.RouterGroup
	BaseApi
}

func NewCrdApi(router *gin.RouterGroup) *CrdApi {
	api := CrdApi{router: router}
	api.Router()
	return &api
}

func (api *CrdApi) Router() {
	api.router.GET("/list", api.List)
	api.router.GET("", api.Detail)
	api.router.DELETE("", middleware.Blocker(), api.Delete)
}

// List CRD 概要列表(分页); 完整定义通过 /resource/yaml 按需查看
func (api *CrdApi) List(ctx *gin.Context) {
	var query dto.CrdPageQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	items, err := service.ListCrds(ctx.Request.Context(), query.ClusterId, query.Refresh)
	if err != nil {
		logs.Error("查询CRD列表失败: %s %v", query.ClusterId, err)
		api.InternalError(ctx, "查询失败", err)
		return
	}
	api.SuccessWithData(ctx, BuildPageResponse(service.FilterCrds(items, query.Keyword), query.Page, query.PageSize))
}

// Detail 单个 CRD 的概要(名称、组、版本、作用域等, 不含 schema)
func (api *CrdApi) Detail(ctx *gin.Context) {
	var base dto.CrdBaseDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	item, err := service.GetCrd(ctx.Request.Context(), base.ClusterId, base.Name)
	if err != nil {
		api.InternalError(ctx, "获取失败", err)
		return
	}
	api.SuccessWithData(ctx, item)
}

// Delete 删除 CRD(连带其下所有自定义资源, 危险操作)
func (api *CrdApi) Delete(ctx *gin.Context) {
	var base dto.CrdBaseDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	if err := service.DeleteCrd(ctx.Request.Context(), base.ClusterId, base.Name); err != nil {
		logs.Error("删除CRD失败: %s %v", base.Name, err)
		api.InternalError(ctx, "删除失败", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}
