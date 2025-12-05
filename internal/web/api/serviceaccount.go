package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

type ServiceAccountApi struct {
	router *gin.RouterGroup
	srv    *service.ServiceAccountService
	BaseApi
}

func NewServiceAccountApi(router *gin.RouterGroup) *ServiceAccountApi {
	return &ServiceAccountApi{
		router: router,
		srv:    service.GetSingletonServiceAccountService(),
	}
}

func (api *ServiceAccountApi) Router() {
	api.router.GET("/list", api.List)
	api.router.POST("/add", middleware.Blocker(), api.Add)
	api.router.PUT("", middleware.Blocker(), api.Update)
	api.router.DELETE("", middleware.Blocker(), api.Delete)
	api.router.POST("/apply", middleware.Blocker(), api.Apply)
}

// List - 查询ServiceAccount列表
func (a *ServiceAccountApi) List(ctx *gin.Context) {
	var query dto.ServiceAccountPageQueryBaseDTO
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

// Add - 创建ServiceAccount
func (api *ServiceAccountApi) Add(ctx *gin.Context) {
	var create dto.ServiceAccountCreateDTO
	if err := ctx.ShouldBindJSON(&create); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Save(create)
	if err != nil {
		api.InternalError(ctx, "创建失败:", err)
		return
	}
	api.SuccessMsg(ctx, "创建成功")
}

// Update - 修改ServiceAccount
func (api *ServiceAccountApi) Update(ctx *gin.Context) {
	var update dto.ServiceAccountCreateDTO
	if err := ctx.ShouldBindJSON(&update); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.Update(update)
	if err != nil {
		api.InternalError(ctx, "修改失败:", err)
		return
	}
	api.SuccessMsg(ctx, "修改成功")
}

// Delete - 删除ServiceAccount
func (api *ServiceAccountApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
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

// Apply - 通过YAML应用ServiceAccount
func (api *ServiceAccountApi) Apply(ctx *gin.Context) {
	var apply dto.BaseYamlApplyDTO
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
