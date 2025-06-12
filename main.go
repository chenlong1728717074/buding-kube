// @title buding-kube
// @version 1.0
// @description  buding-kube api文档
// @host localhost:8888
// @BasePath /api
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"buding-kube/internal"
	_ "buding-kube/pkg/kube"
	"buding-kube/pkg/logs"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// swag init -d ./ -o ./pkg/docs
func main() {
	app := internal.NewApp()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.Start()
	logs.Info("SERVER START SUCCESS")

	// 处理信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logs.Info("Shutdown Server ...")
	// 创建关闭超时上下文
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	// 停止Web应用
	app.Stop(shutdownCtx)

	logs.Info("Server exiting")
}
