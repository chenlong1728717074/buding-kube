package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"

	"github.com/gin-gonic/gin"
)

// DaemonSetApi DaemonSet 列表; 其他操作统一使用 /workload 接口
type DaemonSetApi struct {
	router *gin.RouterGroup
	srv    *service.DaemonSetService
	BaseApi
}

func NewDaemonSetApi(router *gin.RouterGroup) *DaemonSetApi {
	api := DaemonSetApi{
		router: router,
		srv:    service.GetSingletonDaemonSetService(),
	}
	api.Router()
	return &api
}

func (api *DaemonSetApi) Router() {
	api.router.GET("/list", api.List)
}

func (api *DaemonSetApi) List(ctx *gin.Context) {
	var query dto.WorkloadQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	list, err := api.srv.List(query)
	if err != nil {
		api.InternalError(ctx, "查询失败", err)
		return
	}
	api.SuccessWithData(ctx, BuildPageResponse(list, query.Page, query.PageSize))
}
