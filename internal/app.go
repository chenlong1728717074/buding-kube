package internal

import (
	"buding-kube/internal/web/middleware"
	"buding-kube/internal/web/router"
	"buding-kube/pkg/config"
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
	engine := gin.Default()

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
	var port int
	var addr string
	var server *http.Server
	var err error

	//初始化
	port = getPort()
	addr = fmt.Sprintf(":%d", port)
	server = &http.Server{
		Addr:           addr,
		Handler:        app.engine,
		ReadTimeout:    10 * time.Second, // 添加超时配置
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	app.server = server
	//启动
	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logs.Fatal("Server Shutdown err:", err)
		}
	}()
	app.server = server
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

func getPort() int {
	//配置
	var port int
	port = config.GetConfig().Server.Port

	//env
	portStr := os.Getenv("KUBE_RUNTIME_PORT")
	envPort, err := strconv.Atoi(portStr)
	if err == nil {
		port = envPort
	}

	if port <= 0 || port > 65535 {
		logs.Warn("Invalid port %s, using default 8080", portStr)
		return 8080
	}
	return port
}
