package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type CronJobApi struct {
	router *gin.RouterGroup
	srv    *service.CronJobService
	BaseApi
}

func NewCronJobApi(router *gin.RouterGroup) *CronJobApi {
	api := CronJobApi{
		router: router,
		srv:    service.GetSingletonCronJobService(),
	}
	api.Router()
	return &api
}

func (a *CronJobApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *CronJobApi) List(ctx *gin.Context) {
	var query dto.CronJobPageQueryBaseDTO
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
