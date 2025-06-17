package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"buding-kube/pkg/logs"
	"bufio"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type PodApi struct {
	router *gin.RouterGroup
	srv    *service.PodService
	BaseApi
}

func NewPodApi(router *gin.RouterGroup) *PodApi {
	api := PodApi{
		router: router,
		srv:    service.GetSingletonPodService(),
	}
	api.Router()
	return &api
}

func (api *PodApi) Router() {
	api.router.POST("", middleware.Blocker(), api.Add)
	api.router.DELETE("", middleware.Blocker(), api.Delete)
	api.router.PUT("", middleware.Blocker(), api.Update)
	api.router.GET("/list", api.List)
	api.router.POST("/logs", api.Log)
	api.router.GET("", api.Info)
}

func (api *PodApi) Add(ctx *gin.Context) {
	api.InternalError(ctx, "操作失败:", errors.New("pod暂不开放添加和修改能力"))
}

func (api *PodApi) Update(ctx *gin.Context) {
	api.InternalError(ctx, "操作失败:", errors.New("pod暂不开放添加和修改能力"))
}

func (api *PodApi) Delete(ctx *gin.Context) {
	var pod dto.PodDTO
	if err := api.BindQuery(ctx, &pod); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	if err := api.srv.Delete(pod); err != nil {
		api.InternalError(ctx, "删除失败", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}

func (api *PodApi) List(ctx *gin.Context) {
	var query dto.PodQueryDTO
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

func (api *PodApi) Info(ctx *gin.Context) {
	var query dto.PodDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.GetById(query)
	if err != nil {
		api.InternalError(ctx, "获取失败:", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

func (api *PodApi) Log(ctx *gin.Context) {
	var query dto.PodLogDTO
	if err := api.BindJSON(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	stream, err := api.srv.PodLog(query)
	defer stream.Close()
	if err != nil {
		api.InternalError(ctx, "获取失败:", err)
		return
	}
	//升级sse
	api.keepAliveTextPlain(ctx)
	//发送数据
	api.sseListen(ctx, stream, query)
}
func (api *PodApi) sseListen(ctx *gin.Context, stream io.ReadCloser, query dto.PodLogDTO) {
	plan := NewStreamPlan(ctx)
	plan.Writer("Connected successfully\n")
	// 发送开始消息（修复格式）
	plan.Writer(fmt.Sprintf("[START] Pod log streaming started for %s %s %s\n",
		query.ClusterId, query.Namespace, query.Name))

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	dataChan := make(chan string, 10)
	errorChan := make(chan error, 1)

	// 启动读取 goroutine
	scanner := bufio.NewScanner(stream)
	go func() {
		defer func() {
			close(dataChan)
			// 安全关闭 errorChan
			select {
			case errorChan <- nil: // 发送结束信号
			default:
			}
		}()

		for scanner.Scan() {
			select {
			case dataChan <- scanner.Text():
			case <-ctx.Request.Context().Done():
				return
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case errorChan <- err:
			case <-ctx.Request.Context().Done():
			}
		}
	}()

	for {
		select {
		case <-ctx.Request.Context().Done():
			logs.Info("[STOP] Pod log streaming stopped for %s %s %s",
				query.ClusterId, query.Namespace, query.Name)
			return

		case <-heartbeat.C:
			plan.Writer("ping\n")

		case line, ok := <-dataChan:
			if !ok {
				plan.Writer("[END] Log stream ended\n")
				return
			}
			plan.Writer(line + "\n")

		case err := <-errorChan:
			if err != nil {
				plan.Writer(fmt.Sprintf("[ERROR] %s\n", err.Error()))
			}
			return
		}
	}
}
