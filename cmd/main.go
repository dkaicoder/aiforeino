package main

import (
	"context"

	"main/agent"
	"main/config"
	"main/graph/export_graph"
	"main/internal/database"
	"main/internal/observability"
	"main/internal/repository"
	"main/router"

	_ "net/http/pprof"
)

func main() {
	ctx := context.Background()
	config.InitConfig()
	configs := config.C
	observability.InitLangfuseEino(configs)

	database.Init(configs)
	database.InitRedis(ctx)
	database.InitMysql(ctx)
	chatHistoryRepo := repository.NewRedisChatHistoryRepo(database.RedisDb)
	downloadListRepo := repository.NewDownloadListRepo(database.MysqlDb)
	exportGraph := export_graph.NewExportGraph(downloadListRepo)
	agentApi := agent.NewAgent(configs, chatHistoryRepo, exportGraph)
	s := router.NewRouter("/agent", agentApi)
	r := router.NewApp(8080, s)
	r.Run()
}
