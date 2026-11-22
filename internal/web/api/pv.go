package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"github.com/gin-gonic/gin"
)

type PVApi struct {
	router *gin.RouterGroup
	srv    *service.PVService
	BaseApi
}

func NewPVApi(router *gin.RouterGroup) *PVApi {
	api := PVApi{
		router: router,
		srv:    service.GetSingletonPVService(),
	}
	api.Router()
	return &api
}

func (a *PVApi) Router() {
	a.router.GET("/list", a.List)
	a.router.POST("", middleware.Blocker(), BindJSON[dto.PVCreateDTO](a.Create))
	a.router.DELETE("", middleware.Blocker(), a.Delete)
}

// Delete 删除PersistentVolume
func (a *PVApi) Delete(ctx *gin.Context) {
	var base dto.BaseDTO
	if err := a.BindQuery(ctx, &base); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	if err := a.srv.Delete(base); err != nil {
		a.InternalError(ctx, "删除失败:%v", err)
		return
	}
	a.SuccessMsg(ctx, "删除成功")
}

func (a *PVApi) List(ctx *gin.Context) {
	var query dto.PVPageQueryBaseDTO
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

// Create 创建持久卷
func (a *PVApi) Create(ctx *gin.Context, req dto.PVCreateDTO) {
	if err := a.srv.Create(ctx.Request.Context(), req); err != nil {
		a.InternalError(ctx, "创建失败", err)
		return
	}
	a.SuccessMsg(ctx, "创建成功")
}
