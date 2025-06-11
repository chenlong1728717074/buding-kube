package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type NamespacesApi struct {
	router *gin.RouterGroup
	srv    *service.NamespaceService
	BaseApi
}

func NewNamespacesApi(router *gin.RouterGroup) *NamespacesApi {
	api := NamespacesApi{
		router: router,
		srv:    service.GetSingletonNamespaceService(),
	}
	api.Router()
	return &api
}

func (api *NamespacesApi) Router() {
	api.router.GET("/:id", api.Info)
	api.router.DELETE("/:id", api.Delete)
	api.router.GET("/list", api.List)
	api.router.POST("", api.Add)
	api.router.PUT("", api.Update)
}

func (*NamespacesApi) List(ctx *gin.Context) {

}

func (api *NamespacesApi) Info(ctx *gin.Context) {

}

func (api *NamespacesApi) Delete(ctx *gin.Context) {

}

func (api *NamespacesApi) Add(ctx *gin.Context) {
	var create dto.NamespaceCreateDTO
	if err := ctx.ShouldBindJSON(&create); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Save(create)
	if err != nil {
		api.InternalError(ctx, "添加失败:", err)
		return
	}
	api.SuccessMsg(ctx, "添加成功")
}

func (api *NamespacesApi) Update(ctx *gin.Context) {
	var create dto.NamespaceCreateDTO
	if err := ctx.ShouldBindJSON(&create); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Update(create)
	if err != nil {
		api.InternalError(ctx, "修改失败:", err)
		return
	}
	api.SuccessMsg(ctx, "修改成功")

}
