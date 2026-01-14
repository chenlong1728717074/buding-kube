// @title buding-kube
// @version 1.0
// @description  buding-kube api文档
// @host localhost:8888
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"buding-kube/internal"
	_ "buding-kube/internal/kube"
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
	app.Start()
	logs.Info("SERVER STARTED SUCCESSFULLY")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logs.Info("Shutting down server...")

	// 创建关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 停止应用
	if err := app.Stop(ctx); err != nil {
		logs.Error("Failed to stop server gracefully: %v", err)
		os.Exit(1)
	}

	logs.Info("Server exited")
}
