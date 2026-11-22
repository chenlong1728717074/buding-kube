package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"

	"github.com/gin-gonic/gin"
)

type IngressApi struct {
	router *gin.RouterGroup
	srv    *service.IngressService
	BaseApi
}

func NewIngressApi(router *gin.RouterGroup) *IngressApi {
	api := IngressApi{
		router: router,
		srv:    service.GetSingletonIngressService(),
	}
	api.Router()
	return &api
}

func (a *IngressApi) Router() {
	a.router.GET("/list", a.List)
	a.router.GET("", a.Get)
	a.router.GET("/classes", a.Classes)
	a.router.GET("/tlsSecrets", a.TLSSecrets)
	a.router.POST("", middleware.Blocker(), BindJSON[dto.IngressFormDTO](a.Create))
	a.router.PUT("", middleware.Blocker(), BindJSON[dto.IngressFormDTO](a.Update))
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

// Delete 删除Ingress
func (a *IngressApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := a.BindQuery(ctx, &base); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	if err := a.srv.Delete(ctx.Request.Context(), base); err != nil {
		a.InternalError(ctx, "删除失败", err)
		return
	}
	a.SuccessMsg(ctx, "删除成功")
}

func (a *IngressApi) List(ctx *gin.Context) {
	var query dto.IngressPageQueryBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.List(ctx.Request.Context(), query)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, BuildPageResponse(list, query.Page, query.PageSize))
}

// Get 应用路由详情
func (a *IngressApi) Get(ctx *gin.Context) {
	var query dto.IngressBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	detail, err := a.srv.Get(ctx.Request.Context(), query)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, detail)
}

// Classes 集群中的 Ingress 类
func (a *IngressApi) Classes(ctx *gin.Context) {
	var query dto.NamespaceBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.Classes(ctx.Request.Context(), query.ClusterId)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, list)
}

// TLSSecrets 命名空间中的证书
func (a *IngressApi) TLSSecrets(ctx *gin.Context) {
	var query dto.NamespaceBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.TLSSecrets(ctx.Request.Context(), query.ClusterId, query.Namespace)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, list)
}

// Create 创建应用路由
func (a *IngressApi) Create(ctx *gin.Context, req dto.IngressFormDTO) {
	if err := a.srv.Create(ctx.Request.Context(), req, a.CurrentUser(ctx).Username); err != nil {
		a.InternalError(ctx, "创建失败", err)
		return
	}
	a.SuccessMsg(ctx, "创建成功")
}

// Update 编辑应用路由
func (a *IngressApi) Update(ctx *gin.Context, req dto.IngressFormDTO) {
	if err := a.srv.Update(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "更新失败", err)
		return
	}
	a.SuccessMsg(ctx, "更新成功")
}
