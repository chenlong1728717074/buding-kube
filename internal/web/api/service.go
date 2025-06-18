package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
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
}

func (api *KubeSrvApi) List(ctx *gin.Context) {
	var query dto.ServiceQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	list, err := api.srv.List(query)
	if err != nil {
		api.InternalError(ctx, "查询失败:", err)
		return
	}
	response := BuildPageResponse(list, query.Page, query.PageSize)
	api.SuccessWithData(ctx, response)
}
