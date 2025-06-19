package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type PVApi struct {
	router *gin.RouterGroup
	srv    *service.PVService
	BaseApi
}

func NewPVApi(router *gin.RouterGroup) *PVApi {
	api := PVApi{
		router: router,
		srv:    service.GetSingletonPVService(),
	}
	api.Router()
	return &api
}

func (a *PVApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *PVApi) List(ctx *gin.Context) {
	var query dto.PVPageQueryBaseDTO
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
