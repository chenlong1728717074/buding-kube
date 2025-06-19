package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type EndpointSliceApi struct {
	router *gin.RouterGroup
	srv    *service.EndpointSliceService
	BaseApi
}

func NewEndpointSliceApi(router *gin.RouterGroup) *EndpointSliceApi {
	api := &EndpointSliceApi{
		router: router,
		srv:    service.GetSingletonEndpointSliceService(),
	}
	api.Router()
	return api
}

func (a *EndpointSliceApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *EndpointSliceApi) List(ctx *gin.Context) {
	var query dto.EndpointSlicePageQueryBaseDTO
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
