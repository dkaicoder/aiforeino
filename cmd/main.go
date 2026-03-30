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
	"net/http"
	"path/filepath"
	"time"

	_ "net/http/pprof"
)

func main() {
	ctx := context.Background()
	configs := config.InitConfig()
	database.Init(configs)
	redisC := database.InitRedis(ctx)
	db := database.InitMysql(ctx)

	chatHistoryRepo := repository.NewRedisChatHistoryRepo(redisC)
	downloadListRepo := repository.NewDownloadListRepo(db)
	exportGraph := export_graph.NewExportGraph(downloadListRepo)
	agentApi := agent.NewAgent(configs, chatHistoryRepo, exportGraph)

	staticHomeDir, _ := filepath.Abs("./static/home")
	staticDir, _ := filepath.Abs("./static")
	staticHomeFileServer := http.FileServer(http.Dir(staticHomeDir))
	staticFileServer := http.FileServer(http.Dir(staticDir))
	mux := http.NewServeMux()
	mux.Handle("/", staticHomeFileServer)
	mux.Handle("/static/", http.StripPrefix("/static/", staticFileServer))
	mux.HandleFunc("/chat/history", agentApi.GetHis)
	mux.HandleFunc("/stream", agentApi.StreamHandler)
	http.ListenAndServe(":8080", loggingMiddleware(mux))
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
