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
	api.router.GET("", api.Info)
	api.router.DELETE("", api.Delete)
	api.router.GET("/list", api.List)
	api.router.POST("", api.Add)
	api.router.PUT("", api.Update)
	api.router.POST("/apply", api.Apply)
}

func (api *NamespacesApi) List(ctx *gin.Context) {
	var query dto.NamespacePageQueryBaseDTO
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

func (api *NamespacesApi) Info(ctx *gin.Context) {
	var base dto.NamespaceBaseDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.GetById(base)
	if err != nil {
		api.InternalError(ctx, "获取失败:", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

func (api *NamespacesApi) Delete(ctx *gin.Context) {
	var base dto.NamespaceBaseDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	if err := api.srv.Delete(base); err != nil {
		api.InternalError(ctx, "删除失败", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
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

func (api *NamespacesApi) Apply(ctx *gin.Context) {
	var apply dto.NamespaceApplyDTO
	if err := ctx.ShouldBindJSON(&apply); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Apply(apply)
	if err != nil {
		api.InternalError(ctx, "执行失败:", err)
		return
	}
	api.SuccessMsg(ctx, "执行成功")
}
