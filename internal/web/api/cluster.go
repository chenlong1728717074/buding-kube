package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/pkg/logs"
	"github.com/gin-gonic/gin"
)

type ClusterApi struct {
	router *gin.RouterGroup
	srv    *service.ClusterService
	BaseApi
}

func NewClusterApi(router *gin.RouterGroup) *ClusterApi {
	api := ClusterApi{
		router: router,
		srv:    service.GetSingletonClusterService(),
	}
	api.Router()
	return &api
}

func (api *ClusterApi) Router() {
	api.router.GET("/:id", api.Info)
	api.router.DELETE("/:id", api.Delete)
	api.router.GET("/list", api.List)
	api.router.POST("", api.Add)
}
func (api *ClusterApi) Add(ctx *gin.Context) {
	var create dto.NodeCreateDTO
	if err := ctx.ShouldBindJSON(&create); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.SaveOrUpdate(create)
	if err != nil {
		api.InternalError(ctx, "添加失败:", err)
		return
	}
	api.SuccessMsg(ctx, "添加成功")
}

func (api *ClusterApi) List(ctx *gin.Context) {
	var query dto.PageQueryDTO
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

func (api *ClusterApi) Delete(ctx *gin.Context) {
	id := api.GetParam(ctx, "id")
	if id == "" {
		logs.Info("获取不到需集群id")
		api.ParamError(ctx, "获取id失败")
		return
	}
	if err := api.srv.Delete(id); err != nil {
		api.InternalError(ctx, "删除失败:%v", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}
func (api *ClusterApi) Info(ctx *gin.Context) {
	id := api.GetParam(ctx, "id")
	if id == "" {
		logs.Info("获取不到需集群id")
		api.ParamError(ctx, "获取id失败")
		return
	}
	result, err := api.srv.GetById(id)
	if err != nil {
		api.InternalError(ctx, "获取失败:%v", err)
		return
	}
	api.SuccessWithData(ctx, result)
}
