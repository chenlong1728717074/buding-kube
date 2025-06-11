package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
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

func (*ClusterApi) List(ctx *gin.Context) {

}

func (api *ClusterApi) Delete(ctx *gin.Context) {
	id := api.GetParam(ctx, "id")
	if id == "" {
		api.ParamError(ctx, "获取id失败")
		return
	}
	if err := api.srv.Delete(id); err != nil {
		api.InternalError(ctx, "删除失败:", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}
func (*ClusterApi) Info(ctx *gin.Context) {

}
