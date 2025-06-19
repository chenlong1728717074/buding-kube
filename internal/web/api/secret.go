package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type SecretApi struct {
	router *gin.RouterGroup
	srv    *service.SecretService
	BaseApi
}

func NewSecretApi(router *gin.RouterGroup) *SecretApi {
	api := SecretApi{
		router: router,
		srv:    service.GetSingletonSecretService(),
	}
	api.Router()
	return &api
}

func (a *SecretApi) Router() {
	a.router.GET("/list", a.List)
}

func (a *SecretApi) List(ctx *gin.Context) {
	var query dto.SecretPageQueryBaseDTO
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
