package internal

import (
	"buding-kube/internal/web/middleware"
	"buding-kube/internal/web/router"
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
)

type App struct {
	engine *gin.Engine
	server *http.Server
}

func (app *App) Start() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8888),
		Handler: app.engine,
	}
	//路由
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logs.Fatal("Server Shutdown err:", err)
		}
	}()
	app.server = server
}

func (app *App) Stop(ctx context.Context) {
	if err := app.server.Shutdown(ctx); err != nil {
		logs.Error("Server Shutdown err:", err)
	}
}

func NewApp() *App {
	gin.SetMode(gin.DebugMode)
	//日志
	gin.DefaultWriter = logs.NewGinLoggerAdapter(zapcore.InfoLevel)
	gin.DefaultErrorWriter = logs.NewGinLoggerAdapter(zapcore.ErrorLevel)
	engine := gin.Default()
	//404
	engine.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "资源未找到")
	})
	//中间件
	engine.Use(middleware.Cors(), middleware.Logger(), middleware.Recovery())
	//路由
	api := engine.Group("/api")
	router.SetupRouter(api)
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return &App{
		engine: engine,
	}
}
