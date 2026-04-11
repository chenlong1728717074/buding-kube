package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
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
	api.router.GET("/:name", middleware.Blocker(), api.Info)
	api.router.DELETE("/:name", middleware.Blocker(), api.Delete)
	api.router.PUT("/:name", middleware.Blocker(), BindJSON[dto.NodeUpdateDTO](api.Update))
	api.router.GET("/list", api.List)
	api.router.POST("", middleware.Blocker(), BindJSON[dto.NodeCreateDTO](api.Add))
}
func (api *ClusterApi) Add(ctx *gin.Context, create dto.NodeCreateDTO) {
	err := api.srv.SaveOrUpdate(dto.NodeUpdateDTO{
		Name:     create.Name,
		Alias:    create.Alias,
		Describe: create.Describe,
		Config:   create.Config,
		Uri:      create.Uri,
		Token:    create.Token,
	})
	if err != nil {
		api.InternalError(ctx, "添加失败:", err)
		return
	}
	api.SuccessMsg(ctx, "添加成功")
}

func (api *ClusterApi) Update(ctx *gin.Context, update dto.NodeUpdateDTO) {
	name := api.GetParam(ctx, "name")
	if name == "" {
		api.ParamError(ctx, "获取name失败")
		return
	}
	if update.Name == "" {
		update.Name = name
	}
	if update.Name != name {
		api.ParamError(ctx, "路径name与请求体name不一致")
		return
	}
	if err := api.srv.SaveOrUpdate(update); err != nil {
		api.InternalError(ctx, "更新失败:", err)
		return
	}
	api.SuccessMsg(ctx, "更新成功")
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
	name := api.GetParam(ctx, "name")
	if name == "" {
		logs.Info("获取不到需集群name")
		api.ParamError(ctx, "获取name失败")
		return
	}
	if err := api.srv.DeleteByName(name); err != nil {
		api.InternalError(ctx, "删除失败:%v", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}
func (api *ClusterApi) Info(ctx *gin.Context) {
	name := api.GetParam(ctx, "name")
	if name == "" {
		logs.Info("获取不到需集群name")
		api.ParamError(ctx, "获取name失败")
		return
	}
	result, err := api.srv.GetByName(name)
	if err != nil {
		api.InternalError(ctx, "获取失败:%v", err)
		return
	}
	api.SuccessWithData(ctx, result)
}
