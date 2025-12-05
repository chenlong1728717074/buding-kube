package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

type ConfigMapApi struct {
	router *gin.RouterGroup
	srv    *service.ConfigMapService
	BaseApi
}

func NewConfigMapApi(router *gin.RouterGroup) *ConfigMapApi {
	api := ConfigMapApi{
		router: router,
		srv:    service.GetSingletonConfigMapService(),
	}
	api.Router()
	return &api
}

func (api *ConfigMapApi) Router() {
	api.router.GET("/list", api.List)
	api.router.GET("/add", api.Add)
	api.router.POST("/apply", middleware.Blocker(), api.Apply)
	api.router.DELETE("", middleware.Blocker(), api.Delete)
	api.router.PUT("", middleware.Blocker(), api.Update)
	api.router.PUT("/data", middleware.Blocker(), api.UpdateData)
}

func (a *ConfigMapApi) List(ctx *gin.Context) {
	var query dto.ConfigMapPageQueryBaseDTO
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

func (api *ConfigMapApi) Apply(ctx *gin.Context) {
	var apply dto.BaseYamlApplyDTO
	if err := ctx.ShouldBindJSON(&apply); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.BaseYamlApplyDTO(apply)
	if err != nil {
		api.InternalError(ctx, "执行失败:", err)
		return
	}
	api.SuccessMsg(ctx, "执行成功")
}

func (api *ConfigMapApi) Delete(ctx *gin.Context) {
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

// Add - 添加ConfigMap
func (api *ConfigMapApi) Add(ctx *gin.Context) {
	var create dto.ConfigMapCreateDTO
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

// Update - 只允许修改别名和描述
func (api *ConfigMapApi) Update(ctx *gin.Context) {
	var update dto.BaseInfoDTO
	if err := ctx.ShouldBindJSON(&update); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.UpdateInfo(update)
	if err != nil {
		api.InternalError(ctx, "修改失败:", err)
		return
	}
	api.SuccessMsg(ctx, "修改成功")
}

// UpdateData - 只修改data
func (api *ConfigMapApi) UpdateData(ctx *gin.Context) {
	var updateData dto.ConfigMapDataDTO
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.UpdateData(updateData)
	if err != nil {
		api.InternalError(ctx, "修改数据失败:", err)
		return
	}
	api.SuccessMsg(ctx, "修改数据成功")
}
