package main

import (
	"context"
	"log"
	"main/config"
	"main/internal/database"
	"main/internal/kafka"
	"main/internal/repository"
	exports "main/internal/service/export"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config.InitConfig()
	database.Init(config.C)
	database.InitMysql(ctx)

	downLoadRepo := repository.NewDownloadListRepo(database.MysqlDb)
	export := exports.NewExportService(database.MysqlDb, downLoadRepo)

	kafka.StartExportWorkers(ctx, 5, export)

	go kafka.Consumer(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("收到退出信号，开始优雅关闭...")

	cancel()

	close(kafka.ExportChan)

	kafka.ExportWg.Wait()
	log.Println("所有导出协程已退出，程序正常结束")
}
