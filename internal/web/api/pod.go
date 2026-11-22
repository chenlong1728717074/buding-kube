package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"buding-kube/pkg/logs"
	"bufio"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// logHeartbeatInterval 日志流心跳间隔, 防止网关因长时间无数据断开连接
	logHeartbeatInterval = 30 * time.Second
	// logReadBufferSize 日志读取缓冲区; 按行读取不受 bufio.Scanner 64KB 单行上限的限制
	logReadBufferSize = 64 * 1024
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
	api.router.POST("", middleware.Blocker(), BindJSON[dto.PodCreateDTO](api.Create))
	api.router.DELETE("", middleware.Blocker(), api.Delete)
	api.router.PUT("", middleware.Blocker(), BindJSON[dto.PodInfoDTO](api.UpdateInfo))
	api.router.PUT("/image", middleware.Blocker(), BindJSON[dto.PodImageDTO](api.UpdateImages))
	api.router.GET("/list", api.List)
	api.router.POST("/logs", api.Log)
	api.router.POST("/download", middleware.Blocker(), api.Download)
	api.router.POST("/upload", middleware.Blocker(), api.Upload)
	api.router.GET("", api.Info)
	api.router.GET("/expel", api.Expel)
}

// Expel 驱逐pod
func (api *PodApi) Expel(ctx *gin.Context) {
	var query dto.PodDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	if err := api.srv.Expel(query); err != nil {
		api.InternalError(ctx, "驱逐失败", err)
		return
	}
	api.SuccessMsg(ctx, "驱逐成功")
}

// Create 创建容器组
func (api *PodApi) Create(ctx *gin.Context, req dto.PodCreateDTO) {
	if err := api.srv.Create(ctx.Request.Context(), req, api.CurrentUser(ctx).Username); err != nil {
		api.InternalError(ctx, "创建失败", err)
		return
	}
	api.SuccessMsg(ctx, "创建成功")
}

// UpdateInfo 编辑别名和描述
func (api *PodApi) UpdateInfo(ctx *gin.Context, req dto.PodInfoDTO) {
	if err := api.srv.UpdateInfo(ctx.Request.Context(), req); err != nil {
		api.InternalError(ctx, "修改失败", err)
		return
	}
	api.SuccessMsg(ctx, "修改成功")
}

// UpdateImages 更新镜像
func (api *PodApi) UpdateImages(ctx *gin.Context, req dto.PodImageDTO) {
	if err := api.srv.UpdateImages(ctx.Request.Context(), req); err != nil {
		api.InternalError(ctx, "更新镜像失败", err)
		return
	}
	api.SuccessMsg(ctx, "镜像已更新")
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
		api.InternalError(ctx, "查询失败", err)
		return
	}
	api.SuccessWithData(ctx, BuildPageResponse(list, query.Page, query.PageSize))
}

func (api *PodApi) Info(ctx *gin.Context) {
	var query dto.PodDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.Get(ctx.Request.Context(), query)
	if err != nil {
		api.InternalError(ctx, "获取失败", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

// extractFileName 从文件路径中提取文件名
func (api *PodApi) extractFileName(filePath string) string {
	name := path.Base(strings.ReplaceAll(strings.TrimSpace(filePath), "\\", "/"))
	if name == "." || name == "/" || name == ".." {
		return "download"
	}
	return name
}

// setDownloadHeaders 设置文件下载的响应头
func (api *PodApi) setDownloadHeaders(ctx *gin.Context, fileName string) {
	// FormatMediaType 会正确转义引号, 并对中文文件名使用 RFC 2231 编码(filename*=utf-8''...)
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": fileName + ".tar"})
	if disposition == "" {
		disposition = `attachment; filename="download.tar"`
	}
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", disposition)
	ctx.Header("Content-Type", "application/x-tar")
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("X-Accel-Buffering", "no")
}

func (api *PodApi) Download(ctx *gin.Context) {
	var query dto.PodDownloadDTO
	if err := api.BindJSON(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}

	// 传入请求的 Context: 客户端取消下载时, 容器内的 tar 进程会被同步终止
	stream, err := api.srv.Download(ctx.Request.Context(), query)
	if err != nil {
		api.InternalError(ctx, "下载文件失败", err)
		return
	}
	defer stream.Close()

	api.setDownloadHeaders(ctx, api.extractFileName(query.FilePath))
	if _, err = io.Copy(ctx.Writer, stream); err != nil {
		if ctx.Request.Context().Err() != nil {
			logs.Info("客户端取消下载: %s", query.FilePath)
			return
		}
		logs.Error("文件传输失败: %s %v", query.FilePath, err)
		// 还没写出任何内容, 仍可以返回 JSON 错误
		if !ctx.Writer.Written() {
			ctx.Writer.Header().Del("Content-Disposition")
			ctx.Writer.Header().Del("Content-Type")
			api.InternalError(ctx, "文件传输失败", err)
			return
		}
		// 已经写出部分内容, 再追加 JSON 会损坏 tar 文件; 直接中断连接, 让浏览器感知下载失败
		panic(http.ErrAbortHandler)
	}
	logs.Info("文件下载成功: %s", query.FilePath)
}

func (api *PodApi) Upload(ctx *gin.Context) {
	var query dto.PodDownloadDTO
	if err := api.BindForm(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	// 超过 MaxMultipartMemory 的部分由 Go 暂存在磁盘临时文件中, 这里拿到的是文件句柄, 后续流式读取
	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		api.ParamError(ctx, "获取上传文件失败: "+err.Error())
		return
	}
	defer file.Close()
	if err = api.srv.UploadWithTar(ctx.Request.Context(), query, file, fileHeader); err != nil {
		api.InternalError(ctx, "上传文件失败", err)
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
	stream, err := api.srv.PodLog(ctx.Request.Context(), query)
	if err != nil {
		api.InternalError(ctx, "获取失败", err)
		return
	}
	defer stream.Close()
	// 升级为流式响应
	api.keepAliveTextPlain(ctx)
	api.streamLogs(ctx, stream, query)
}

// streamLogs 逐行转发日志, 并定时发送心跳
func (api *PodApi) streamLogs(ctx *gin.Context, stream io.Reader, query dto.PodLogDTO) {
	plan := NewStreamPlan(ctx)
	plan.Writer("Connected successfully\n")
	plan.Writer(fmt.Sprintf("[START] Pod log streaming started for %s %s %s\n",
		query.ClusterId, query.Namespace, query.Name))

	reqCtx := ctx.Request.Context()
	lines := make(chan string, 64)
	// readErr 在 close(lines) 之前写入, 主循环在 lines 关闭之后才读取, 不存在数据竞争
	var readErr error
	go func() {
		defer close(lines)
		reader := bufio.NewReaderSize(stream, logReadBufferSize)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				select {
				case lines <- strings.TrimRight(line, "\r\n"):
				case <-reqCtx.Done():
					return
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					readErr = err
				}
				return
			}
		}
	}()

	heartbeat := time.NewTicker(logHeartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-reqCtx.Done():
			logs.Info("[STOP] Pod log streaming stopped for %s %s %s",
				query.ClusterId, query.Namespace, query.Name)
			return

		case <-heartbeat.C:
			plan.Writer("ping\n")

		case line, ok := <-lines:
			if !ok {
				if readErr != nil {
					plan.Writer(fmt.Sprintf("[ERROR] %s\n", readErr.Error()))
				} else {
					plan.Writer("[END] Log stream ended\n")
				}
				return
			}
			plan.Writer(line + "\n")
		}
	}
}
