package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type ConfigMapApi struct {
	router *gin.RouterGroup
	srv    *service.ConfigMapService
	BaseApi
}

func NewConfigMapApi(router *gin.RouterGroup) *ConfigMapApi {
	api := ConfigMapApi{
		router: router,
		srv:    service.GetSingletonConfigMapService(),
	}
	api.Router()
	return &api
}

func (a *ConfigMapApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *ConfigMapApi) List(ctx *gin.Context) {
	var query dto.ConfigMapPageQueryBaseDTO
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
