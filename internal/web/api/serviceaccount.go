package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type ServiceAccountApi struct {
	router *gin.RouterGroup
	srv    *service.ServiceAccountService
	BaseApi
}

func NewServiceAccountApi(router *gin.RouterGroup) *ServiceAccountApi {
	return &ServiceAccountApi{
		router: router,
		srv:    service.GetSingletonServiceAccountService(),
	}
}

func (a *ServiceAccountApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *ServiceAccountApi) List(ctx *gin.Context) {
	var query dto.ServiceAccountPageQueryBaseDTO
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
