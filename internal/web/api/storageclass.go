package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

type StorageClassApi struct {
	router *gin.RouterGroup
	srv    *service.StorageClassService
	BaseApi
}

func NewStorageClassApi(router *gin.RouterGroup) *StorageClassApi {
	api := StorageClassApi{
		router: router,
		srv:    service.GetSingletonStorageClassService(),
	}
	api.Router()
	return &api
}

func (a *StorageClassApi) Router() {
	a.router.GET("/list", a.List)
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

// Delete 删除StorageClass
func (a *StorageClassApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := a.BindQuery(ctx, &base); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	if err := a.srv.Delete(base); err != nil {
		a.InternalError(ctx, "删除失败:%v", err)
		return
	}
	a.SuccessMsg(ctx, "删除成功")
}

func (a *StorageClassApi) List(ctx *gin.Context) {
	var query dto.StorageClassPageQueryBaseDTO
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
