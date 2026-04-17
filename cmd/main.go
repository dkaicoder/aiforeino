package main

import (
	"context"
	"fmt"
	"main/agent"
	"main/agent/tool/rag"
	"main/config"
	"main/graph/export_graph"
	"main/internal/database"
	"main/internal/repository"
	"main/router"
	"net/http"
	"time"

	_ "net/http/pprof"
)

func main() {
	ctx := context.Background()
	config.InitConfig()
	configs := config.C
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %s %s body:%s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path, r.URL.RawQuery)
		next.ServeHTTP(w, r)
	})
}

func Run(ctx context.Context, filePath string) {
	wf := rag.BuildWorkflow(ctx)
	runner, err := wf.Compile(ctx)
	if err != nil {
		fmt.Println(err)
	}
	out, err := runner.Invoke(ctx, rag.Input{FilePath: filePath})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(out)
}
