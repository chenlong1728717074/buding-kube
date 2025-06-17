package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

type NodeApi struct {
	router *gin.RouterGroup
	srv    *service.NodeService
	BaseApi
}

func NewNodeApi(router *gin.RouterGroup) *NodeApi {
	api := NodeApi{
		router: router,
		srv:    service.GetSingletonNodeService(),
	}
	api.Router()
	return &api
}

func (api *NodeApi) Router() {
	api.router.GET("/list", api.List)
	api.router.GET("", api.Info)
	api.router.PUT("/unSchedule", middleware.Blocker(), api.UnSchedule)
}

func (api *NodeApi) List(ctx *gin.Context) {
	var query dto.NodeQueryDTO
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

func (api *NodeApi) Info(ctx *gin.Context) {
	var base dto.NodeInfoQueryDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.GetNodeInfo(base)
	if err != nil {
		api.InternalError(ctx, "获取失败:", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

func (api *NodeApi) UnSchedule(ctx *gin.Context) {
	var query dto.NodeUnScheduleDTO
	if err := api.BindJSON(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.UnSchedule(query)
	if err != nil {
		api.InternalError(ctx, "操作失败:", err)
		return
	}
	api.SuccessMsg(ctx, "操作成功")
}
