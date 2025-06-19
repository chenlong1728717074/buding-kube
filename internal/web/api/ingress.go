package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type IngressApi struct {
	router *gin.RouterGroup
	srv    *service.IngressService
	BaseApi
}

func NewIngressApi(router *gin.RouterGroup) *IngressApi {
	api := IngressApi{
		router: router,
		srv:    service.GetSingletonIngressService(),
	}
	api.Router()
	return &api
}

func (a *IngressApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *IngressApi) List(ctx *gin.Context) {
	var query dto.IngressPageQueryBaseDTO
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
