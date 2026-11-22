package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

type PVCApi struct {
	router *gin.RouterGroup
	srv    *service.PVCService
	BaseApi
}

func NewPVCApi(router *gin.RouterGroup) *PVCApi {
	api := PVCApi{
		router: router,
		srv:    service.GetSingletonPVCService(),
	}
	api.Router()
	return &api
}

func (a *PVCApi) Router() {
	a.router.GET("/list", a.List)
	a.router.POST("", middleware.Blocker(), BindJSON[dto.PVCCreateDTO](a.Create))
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

// Delete 删除PersistentVolumeClaim
func (a *PVCApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := a.BindQuery(ctx, &base); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	if err := a.srv.Delete(ctx.Request.Context(), base); err != nil {
		a.InternalError(ctx, "删除失败:%v", err)
		return
	}
	a.SuccessMsg(ctx, "删除成功")
}

func (a *PVCApi) List(ctx *gin.Context) {
	var query dto.PVCPageQueryBaseDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	list, err := a.srv.List(ctx.Request.Context(), query)
	if err != nil {
		a.InternalError(ctx, "查询失败:", err)
		return
	}
	response := BuildPageResponse(list, query.Page, query.PageSize)
	a.SuccessWithData(ctx, response)
}

// Create 创建持久卷声明
func (a *PVCApi) Create(ctx *gin.Context, req dto.PVCCreateDTO) {
	if err := a.srv.Create(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "创建失败", err)
		return
	}
	a.SuccessMsg(ctx, "创建成功")
}
