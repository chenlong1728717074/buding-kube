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
	api.router.GET("/usage", api.Usage)
	api.router.GET("", api.Info)
	api.router.PUT("/unSchedule", middleware.Blocker(), api.UnSchedule)
	api.router.PUT("/labels", middleware.Blocker(), BindJSON[dto.NodeLabelsDTO](api.UpdateLabels))
	api.router.PUT("/taints", middleware.Blocker(), BindJSON[dto.NodeTaintsDTO](api.UpdateTaints))
	api.router.GET("/drain/preview", api.DrainPreview)
	api.router.POST("/drain", middleware.Blocker(), BindJSON[dto.NodeDrainDTO](api.Drain))
	api.router.DELETE("", middleware.Blocker(), api.Delete)
}

// UpdateLabels 设置节点标签
func (api *NodeApi) UpdateLabels(ctx *gin.Context, req dto.NodeLabelsDTO) {
	if err := api.srv.UpdateLabels(ctx.Request.Context(), req); err != nil {
		api.InternalError(ctx, "保存失败", err)
		return
	}
	api.SuccessMsg(ctx, "标签已更新")
}

// UpdateTaints 设置节点污点
func (api *NodeApi) UpdateTaints(ctx *gin.Context, req dto.NodeTaintsDTO) {
	if err := api.srv.UpdateTaints(ctx.Request.Context(), req); err != nil {
		api.InternalError(ctx, "保存失败", err)
		return
	}
	api.SuccessMsg(ctx, "污点已更新")
}

// DrainPreview 排空预览
func (api *NodeApi) DrainPreview(ctx *gin.Context) {
	var query dto.NodeDrainDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.DrainPreview(ctx.Request.Context(), query)
	if err != nil {
		api.InternalError(ctx, "查询失败", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

// Drain 排空节点
func (api *NodeApi) Drain(ctx *gin.Context, req dto.NodeDrainDTO) {
	result, err := api.srv.Drain(ctx.Request.Context(), req, api.CurrentUser(ctx).Username)
	if err != nil {
		api.InternalError(ctx, "排空失败", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

func (api *NodeApi) Usage(ctx *gin.Context) {
	var query dto.NodeUsageQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.GetClusterUsage(query)
	if err != nil {
		api.InternalError(ctx, "获取集群资源使用情况失败:", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

// Delete 删除Node对象(节点重新注册后会再次出现)
func (api *NodeApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	if err := api.srv.Delete(base); err != nil {
		api.InternalError(ctx, "删除失败:%v", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
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
