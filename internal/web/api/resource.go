package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"errors"

	"github.com/gin-gonic/gin"
)

// ResourceApi 与具体资源类型无关的通用接口
type ResourceApi struct {
	router *gin.RouterGroup
	BaseApi
}

func NewResourceApi(router *gin.RouterGroup) *ResourceApi {
	api := ResourceApi{router: router}
	api.Router()
	return &api
}

func (api *ResourceApi) Router() {
	api.router.GET("/yaml", api.Yaml)
	api.router.PUT("/yaml", middleware.Blocker(), BindJSON[dto.ResourceYamlApplyDTO](api.UpdateYaml))
	api.router.POST("/apply", middleware.Blocker(), BindJSON[dto.ResourceYamlApplyDTO](api.Apply))
	api.router.GET("/table", api.Table)
	api.router.GET("/object", api.Object)
	api.router.GET("/objects", api.Objects)
	api.router.DELETE("", middleware.Blocker(), api.Delete)
}

func toResourceRef(r dto.ResourceRefDTO) service.ResourceRef {
	return service.ResourceRef{
		Kind:       r.Kind,
		Group:      r.Group,
		Version:    r.Version,
		Resource:   r.Resource,
		KindName:   r.KindName,
		Namespaced: r.Namespaced,
	}
}

// Yaml 查询单个资源的 YAML
// @Router /api/resource/yaml [get]
func (api *ResourceApi) Yaml(ctx *gin.Context) {
	var query dto.ResourceTargetDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	content, err := service.ResourceYaml(ctx.Request.Context(), query.ClusterId, toResourceRef(query.ResourceRefDTO), query.Namespace, query.Name)
	if err != nil {
		api.InternalError(ctx, "获取YAML失败", err)
		return
	}
	api.SuccessWithData(ctx, gin.H{"yaml": content})
}

// UpdateYaml 编辑资源的 YAML 并保存
// @Router /api/resource/yaml [put]
func (api *ResourceApi) UpdateYaml(ctx *gin.Context, req dto.ResourceYamlApplyDTO) {
	if req.Name == "" {
		api.ParamErrorWithError(ctx, errors.New("name 不能为空"))
		return
	}
	err := service.ApplyResourceYaml(ctx.Request.Context(), service.ApplyYamlParams{
		ClusterId: req.ClusterId,
		Ref:       toResourceRef(req.ResourceRefDTO),
		Namespace: req.Namespace,
		Name:      req.Name,
		Yaml:      req.Yaml,
	})
	if err != nil {
		api.InternalError(ctx, "保存失败", err)
		return
	}
	api.SuccessMsg(ctx, "保存成功")
}

// Apply 通过 YAML 创建资源(已存在则更新)
// @Router /api/resource/apply [post]
func (api *ResourceApi) Apply(ctx *gin.Context, req dto.ResourceYamlApplyDTO) {
	err := service.ApplyResourceYaml(ctx.Request.Context(), service.ApplyYamlParams{
		ClusterId:   req.ClusterId,
		Ref:         toResourceRef(req.ResourceRefDTO),
		Namespace:   req.Namespace,
		Yaml:        req.Yaml,
		AllowCreate: true,
		CreateOnly:  req.CreateOnly,
	})
	if err != nil {
		api.InternalError(ctx, "执行失败", err)
		return
	}
	api.SuccessMsg(ctx, "执行成功")
}

// Table 通用列表(与 kubectl get 的列一致)
// @Router /api/resource/table [get]
func (api *ResourceApi) Table(ctx *gin.Context) {
	var query dto.ResourceTableQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := service.ListResourceTable(ctx.Request.Context(), service.ResourceTableQuery{
		ClusterId:     query.ClusterId,
		Ref:           toResourceRef(query.ResourceRefDTO),
		Namespace:     query.Namespace,
		Keyword:       query.Keyword,
		LabelSelector: query.LabelSelector,
		Page:          query.Page,
		PageSize:      query.PageSize,
	})
	if err != nil {
		api.InternalError(ctx, "查询失败", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

// Object 单个资源的完整对象(包含 status)
// @Router /api/resource/object [get]
func (api *ResourceApi) Object(ctx *gin.Context) {
	var query dto.ResourceTargetDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	obj, err := service.GetResourceObject(ctx.Request.Context(), query.ClusterId, toResourceRef(query.ResourceRefDTO), query.Namespace, query.Name)
	if err != nil {
		api.InternalError(ctx, "查询失败", err)
		return
	}
	api.SuccessWithData(ctx, obj)
}

// Objects 列出完整对象
// @Router /api/resource/objects [get]
func (api *ResourceApi) Objects(ctx *gin.Context) {
	var query dto.ResourceObjectsQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	list, err := service.ListResourceObjects(ctx.Request.Context(), query.ClusterId, toResourceRef(query.ResourceRefDTO), query.Namespace, query.LabelSelector)
	if err != nil {
		api.InternalError(ctx, "查询失败", err)
		return
	}
	api.SuccessWithData(ctx, list)
}

// Delete 删除资源
// @Router /api/resource [delete]
func (api *ResourceApi) Delete(ctx *gin.Context) {
	var query dto.ResourceTargetDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	if err := service.DeleteResource(ctx.Request.Context(), query.ClusterId, toResourceRef(query.ResourceRefDTO), query.Namespace, query.Name); err != nil {
		api.InternalError(ctx, "删除失败", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}
