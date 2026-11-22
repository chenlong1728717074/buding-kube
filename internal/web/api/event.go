package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"

	"github.com/gin-gonic/gin"
)

// EventApi 事件中心
type EventApi struct {
	router *gin.RouterGroup
	BaseApi
}

func NewEventApi(router *gin.RouterGroup) *EventApi {
	api := EventApi{router: router}
	api.Router()
	return &api
}

func (a *EventApi) Router() {
	a.router.GET("/list", a.List)
}

// List 事件列表(分页)
func (a *EventApi) List(ctx *gin.Context) {
	var query dto.EventQueryDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	result, err := service.ListEvents(ctx.Request.Context(), query)
	if err != nil {
		a.InternalError(ctx, "查询失败", err)
		return
	}
	a.SuccessWithData(ctx, vo.EventPageVO{
		PageResponse: BuildPageResponse(result.Items, query.Page, query.PageSize),
		Matched:      result.Matched,
		Limited:      result.Limited,
		Limit:        service.EventDefaultLimit,
	})
}
