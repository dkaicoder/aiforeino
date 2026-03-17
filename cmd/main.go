package main

import (
	"context"
	"fmt"
	"main/agent"
	"main/config"
	"main/graph/export_graph"
	"main/internal/database"
	"main/internal/repository"
	"net/http"
	"time"

	_ "net/http/pprof"
)

const (
	prefix = "OuterCyrex:"
	index  = "OuterIndex"
)

func main() {
	ctx := context.Background()
	configs := config.InitConfig()
	database.Init(configs)
	redisC := database.InitRedis(ctx)
	db := database.InitMysql(ctx)
	_ = database.InitKafkaForProducer(ctx)
	chatHistoryRepo := repository.NewRedisChatHistoryRepo(redisC)
	downloadListRepo := repository.NewDownloadListRepo(db)
	exportGraph := export_graph.NewExportGraph(downloadListRepo)
	agentApi := agent.NewAgent(configs, chatHistoryRepo, exportGraph)

	fileServer := http.FileServer(http.Dir("static"))
	mux := http.NewServeMux()
	mux.Handle("/", fileServer)
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

//func save(ctx context.Context) {
//	r, err := demo_rag.InitRAGEngine(ctx, index, prefix)
//	if err != nil {
//		panic(err)
//	}
//
//	doc, err := r.Loader.Load(ctx, document.Source{
//		URI: "./information/mysql-1.md",
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	docs, err := r.Splitter.Transform(ctx, doc)
//	if err != nil {
//		panic(err)
//	}
//
//	for _, d := range docs {
//		uuid, _ := uuid2.NewUUID()
//		d.ID = uuid.String()
//	}
//
//	err = r.InitVectorIndex(ctx)
//	if err != nil {
//		panic(err)
//	}
//
//	_, err = r.Indexer.Store(ctx, docs)
//	if err != nil {
//		panic(err)
//	}
//
//	//var query string
//	//for {
//	//	_, _ = fmt.Scan(&query)
//	//	output, err := r.Generate(ctx, query)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//	var fullContent string // 用来拼接所有片段
//	//	for {
//	//		o, err := output.Recv()
//	//		if err != nil {
//	//			if err == io.EOF {
//	//				break
//	//			}
//	//			panic(err) // 其他错误才 panic
//	//		}
//	//		if o.Content != "" {
//	//			fullContent += o.Content
//	//			fmt.Print(o.Content)
//	//		}
//	//	}
//	//}
//}
