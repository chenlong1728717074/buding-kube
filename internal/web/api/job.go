package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type JobApi struct {
	router *gin.RouterGroup
	srv    *service.JobService
	BaseApi
}

func NewJobApi(router *gin.RouterGroup) *JobApi {
	api := JobApi{
		router: router,
		srv:    service.GetSingletonJobService(),
	}
	api.Router()
	return &api
}

func (a *JobApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *JobApi) List(ctx *gin.Context) {
	var query dto.JobPageQueryBaseDTO
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
