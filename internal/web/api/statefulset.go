package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type StatefulSetApi struct {
	router *gin.RouterGroup
	srv    *service.StatefulSetService
	BaseApi
}

func NewStatefulSetApi(router *gin.RouterGroup) *StatefulSetApi {
	api := StatefulSetApi{
		router: router,
		srv:    service.GetSingletonStatefulSetService(),
	}
	api.Router()
	return &api
}

func (api *StatefulSetApi) Router() {
	api.router.GET("/list", api.List)
}

func (api *StatefulSetApi) List(ctx *gin.Context) {
	var query dto.WorkloadQueryDTO
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
