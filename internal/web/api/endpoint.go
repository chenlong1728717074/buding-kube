package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

// EndpointApi 已过时，建议使用 EndpointSliceApi
// Deprecated: Use EndpointSliceApi instead
type EndpointApi struct {
	router *gin.RouterGroup
	srv    *service.EndpointService
	BaseApi
}

func NewEndpointApi(router *gin.RouterGroup) *EndpointApi {
	api := EndpointApi{
		router: router,
		srv:    service.GetSingletonEndpointService(),
	}
	api.Router()
	return &api
}

func (a *EndpointApi) Router() {
	a.router.GET("/list", a.List)
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

// Delete 删除Endpoint(注意: 由控制器管理, 删除后可能被自动重建)
func (a *EndpointApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := a.BindQuery(ctx, &base); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	if err := a.srv.Delete(ctx.Request.Context(), base); err != nil {
		a.InternalError(ctx, "删除失败:%v", err)
		return
	}
	a.SuccessMsg(ctx, "删除成功")
}

func (a *EndpointApi) List(ctx *gin.Context) {
	var query dto.EndpointPageQueryBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.List(ctx.Request.Context(), query)
	if err != nil {
		a.InternalError(ctx, "查询失败:", err)
		return
	}
	response := BuildPageResponse(list, query.Page, query.PageSize)
	a.SuccessWithData(ctx, response)
}
