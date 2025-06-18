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
	api.router.POST("/download", api.Download)
	api.router.POST("/upload", api.Upload)
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

// extractFileName 从文件路径中提取文件名
func (api *PodApi) extractFileName(filePath string) string {
	if filePath == "" {
		return "download"
	}

	// 找到最后一个路径分隔符
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			return filePath[i+1:]
		}
	}
	return filePath
}

// setDownloadHeaders 设置文件下载的响应头
func (api *PodApi) setDownloadHeaders(ctx *gin.Context, fileName string) {
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.tar", fileName))
	ctx.Header("Content-Type", "application/x-tar")
}

func (api *PodApi) Download(ctx *gin.Context) {
	var query dto.PodDownloadDTO
	if err := api.BindJSON(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}

	// 调用服务层下载文件
	stream, err := api.srv.Download(query)
	if err != nil {
		api.InternalError(ctx, "下载文件失败: ", err)
		return
	}
	defer stream.Close()

	// 提取文件名并设置响应头
	fileName := api.extractFileName(query.FilePath)
	api.setDownloadHeaders(ctx, fileName)

	// 流式传输文件内容
	if _, err = io.Copy(ctx.Writer, stream); err != nil {
		logs.Error("文件传输失败: %v", err)
		api.InternalError(ctx, "文件传输失败: ", err)
		return
	}

	logs.Info("文件下载成功: %s", query.FilePath)
}

func (api *PodApi) Upload(ctx *gin.Context) {
	var query dto.PodDownloadDTO
	if err := api.BindForm(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	file, fileHeader, err := ctx.Request.FormFile("file") // "file" 是前端表单中的字段名
	if err != nil {
		api.ParamError(ctx, "获取上传文件失败: "+err.Error())
		return
	}
	defer file.Close()
	err = api.srv.UploadWithTar(query, file, fileHeader)
	if err != nil {
		api.InternalError(ctx, "上传文件失败: ", err)
		return
	}
	api.SuccessMsg(ctx, "上传成功")
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
