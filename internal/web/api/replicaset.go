package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type ReplicaSetApi struct {
	router *gin.RouterGroup
	srv    *service.ReplicaSetService
	BaseApi
}

func NewReplicaSetApi(router *gin.RouterGroup) *ReplicaSetApi {
	api := ReplicaSetApi{
		router: router,
		srv:    service.GetSingletonReplicaSetService(),
	}
	api.Router()
	return &api
}

func (a *ReplicaSetApi) Router() {
	a.router.GET("/list", a.List)
}

// List 获取ReplicaSet列表
func (a *ReplicaSetApi) List(ctx *gin.Context) {
	var query dto.WorkloadQueryDTO
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
