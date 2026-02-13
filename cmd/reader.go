package main

import (
	"context"
	"log"
	"main/config"
	"main/internal/database"
	"main/internal/kafka"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	configs := config.InitConfig()
	database.Init(configs)
	db := database.InitMysql(ctx)
	export := kafka.NewExportDataBase(db)

	kafka.StartExportWorkers(ctx, 5, export)

	go kafka.KafkaReader(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("收到退出信号，开始优雅关闭...")

	cancel()

	close(kafka.ExportChan)

	kafka.ExportWg.Wait()
	log.Println("所有导出协程已退出，程序正常结束")
}
