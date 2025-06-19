package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type PVCApi struct {
	router *gin.RouterGroup
	srv    *service.PVCService
	BaseApi
}

func NewPVCApi(router *gin.RouterGroup) *PVCApi {
	api := PVCApi{
		router: router,
		srv:    service.GetSingletonPVCService(),
	}
	api.Router()
	return &api
}

func (a *PVCApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *PVCApi) List(ctx *gin.Context) {
	var query dto.PVCPageQueryBaseDTO
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
