package internal

import (
	"buding-kube/internal/web/middleware"
	"buding-kube/internal/web/router"
	"buding-kube/pkg/config"
	"buding-kube/pkg/consts"
	_ "buding-kube/pkg/docs"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type App struct {
	engine *gin.Engine
	server *http.Server
}

func NewApp() *App {
	gin.SetMode(gin.ReleaseMode)
	//日志
	gin.DefaultWriter = logs.NewGinLoggerAdapter(zapcore.InfoLevel)
	gin.DefaultErrorWriter = logs.NewGinLoggerAdapter(zapcore.ErrorLevel)
	// 使用 gin.New 而不是 gin.Default: 访问日志和异常恢复已由自定义中间件负责,
	// gin.Default 会再挂一套 Logger/Recovery, 导致日志重复打印、panic 被重复处理
	engine := gin.New()

	//中间件
	engine.Use(middleware.Cors(), middleware.Logger(), middleware.Recovery())

	//404
	engine.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "资源未找到")
	})

	//路由
	api := engine.Group("/api")
	router.SetupRouter(api)
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &App{engine: engine}
}

func (app *App) Start() {
	addr := fmt.Sprintf(":%d", getPort())
	server := &http.Server{
		Addr:              addr,
		Handler:           app.engine,
		ReadHeaderTimeout: 10 * time.Second, // 限制请求头读取时间, 防御 slowloris
		ReadTimeout:       0,                // 不限制请求体读取: 大文件上传耗时可能远超 10 秒
		WriteTimeout:      0,                // 不限制写超时: Pod 日志、文件下载是长连接流式响应
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	app.server = server
	//启动
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logs.Fatal("Server Shutdown err: %v", err)
		}
	}()
	logs.Info("HTTP 服务监听地址 %s", addr)
}

func (app *App) Stop(ctx context.Context) error {
	if app.server == nil {
		return nil
	}

	if err := app.server.Shutdown(ctx); err != nil {
		logs.Error("Server shutdown error: %v", err)
		return err
	}

	logs.Info("Server stopped successfully")
	return nil
}

// getPort 端口优先级: 环境变量 KUBE_RUNTIME_PORT > 配置文件 server.port > 默认 8080
func getPort() int {
	port := config.GetConfig().Server.Port

	if portStr := strings.TrimSpace(os.Getenv("KUBE_RUNTIME_PORT")); portStr != "" {
		envPort, err := strconv.Atoi(portStr)
		if err != nil {
			logs.Warn("环境变量 KUBE_RUNTIME_PORT=%s 不是合法端口, 已忽略", portStr)
		} else {
			port = envPort
		}
	}

	if port <= 0 || port > 65535 {
		if port != 0 {
			logs.Warn("非法端口 %d, 使用默认端口 %d", port, consts.ServerPort)
		}
		return consts.ServerPort
	}
	return port
}
