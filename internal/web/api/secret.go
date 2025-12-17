package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"

	"github.com/gin-gonic/gin"
)

type SecretApi struct {
	router *gin.RouterGroup
	srv    *service.SecretService
	BaseApi
}

func NewSecretApi(router *gin.RouterGroup) *SecretApi {
	api := SecretApi{
		router: router,
		srv:    service.GetSingletonSecretService(),
	}
	api.Router()
	return &api
}

func (api *SecretApi) Router() {
	api.router.GET("/list", api.List)
	api.router.GET("", api.Get)
	api.router.POST("/add", middleware.Blocker(), api.Add)
	api.router.PUT("", middleware.Blocker(), api.Update)
	api.router.PUT("/data", middleware.Blocker(), api.UpdateData)
	api.router.PUT("/setting", middleware.Blocker(), api.UpdateSetting)
	api.router.POST("/apply", middleware.Blocker(), api.Apply)
	api.router.DELETE("", middleware.Blocker(), api.Delete)
}

// List - 查询Secret列表
func (a *SecretApi) List(ctx *gin.Context) {
	var query dto.ResourcePageQueryBaseDTO
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

func (a *SecretApi) Get(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := a.BindQuery(ctx, &base); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	data, err := a.srv.Get(base)
	if err != nil {
		a.InternalError(ctx, "查询失败:", err)
		return
	}
	a.SuccessWithData(ctx, data)
}

// Add - 创建Secret
func (api *SecretApi) Add(ctx *gin.Context) {
	var create dto.SecretCreateDTO
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

// Update - 只允许修改别名和描述
func (api *SecretApi) Update(ctx *gin.Context) {
	var update dto.SecretCreateDTO
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

// UpdateData - 修改Secret数据
func (api *SecretApi) UpdateData(ctx *gin.Context) {
	var updateData dto.SecretCreateDTO
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

// UpdateSetting - 修改类型、标签、注解
func (api *SecretApi) UpdateSetting(ctx *gin.Context) {
	var update dto.SecretCreateDTO
	if err := ctx.ShouldBindJSON(&update); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	err := api.srv.UpdateSetting(update)
	if err != nil {
		api.InternalError(ctx, "修改设置失败:", err)
		return
	}
	api.SuccessMsg(ctx, "修改设置成功")
}

// Apply - 通过YAML应用Secret
func (api *SecretApi) Apply(ctx *gin.Context) {
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

// Delete - 删除Secret
func (api *SecretApi) Delete(ctx *gin.Context) {
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
