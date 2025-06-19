package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
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
}

func (a *EndpointApi) List(ctx *gin.Context) {
	var query dto.EndpointPageQueryBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.List(query)
	if err != nil {
		a.InternalError(ctx, "查询失败:", err)
		return
	}
	response := BuildPageResponse(list, query.Page, query.PageSize)
	a.SuccessWithData(ctx, response)
}
