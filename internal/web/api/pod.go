package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"errors"
	"github.com/gin-gonic/gin"
)

type PodApi struct {
	router *gin.RouterGroup
	srv    *service.PodService
	BaseApi
}

func NewPodApi(router *gin.RouterGroup) *PodApi {
	api := PodApi{
		router: router,
		srv:    service.GetSingletonPodService(),
	}
	api.Router()
	return &api
}

func (api *PodApi) Router() {
	api.router.POST("", middleware.Blocker(), api.Add)
	api.router.DELETE("", middleware.Blocker(), api.Delete)
	api.router.PUT("", middleware.Blocker(), api.Update)
	api.router.GET("/list", api.List)
	api.router.GET("", api.Info)
}

func (api *PodApi) Add(ctx *gin.Context) {
	api.InternalError(ctx, "操作失败:", errors.New("pod暂不开放添加和修改能力"))
}

func (api *PodApi) Update(ctx *gin.Context) {
	api.InternalError(ctx, "操作失败:", errors.New("pod暂不开放添加和修改能力"))
}

func (api *PodApi) Delete(ctx *gin.Context) {
	var pod dto.PodDTO
	if err := api.BindQuery(ctx, &pod); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	if err := api.srv.Delete(pod); err != nil {
		api.InternalError(ctx, "删除失败", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}

func (api *PodApi) List(ctx *gin.Context) {
	var query dto.PodQueryDTO
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

func (api *PodApi) Info(ctx *gin.Context) {
	var query dto.PodDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.GetById(query)
	if err != nil {
		api.InternalError(ctx, "获取失败:", err)
		return
	}
	api.SuccessWithData(ctx, result)
}
