package main

import (
	"context"
	"main/config"
	"main/internal/database"
	"main/internal/service/export"
	"main/rag_demo"
	"net/http"

	"github.com/cloudwego/eino/components/document"
	uuid2 "github.com/google/uuid"
)

const (
	prefix = "OuterCyrex:"
	index  = "OuterIndex"
)

func main() {
	ctx := context.Background()
	configs := config.InitConfig()
	database.Init(configs)
	database.InitRedis(ctx)
	database.InitMysql(ctx)

	exportService := export.ExportService{}
	fileServer := http.FileServer(http.Dir("static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/chat/history", exportService.GetHis)
	http.HandleFunc("/stream", exportService.StreamHandler)
	http.ListenAndServe(":8080", nil)
}

func save(ctx context.Context) {
	r, err := rag_demo.InitRAGEngine(ctx, index, prefix)
	if err != nil {
		panic(err)
	}

	doc, err := r.Loader.Load(ctx, document.Source{
		URI: "./information/mysql-1.md",
	})
	if err != nil {
		panic(err)
	}

	docs, err := r.Splitter.Transform(ctx, doc)
	if err != nil {
		panic(err)
	}

	for _, d := range docs {
		uuid, _ := uuid2.NewUUID()
		d.ID = uuid.String()
	}

	err = r.InitVectorIndex(ctx)
	if err != nil {
		panic(err)
	}

	_, err = r.Indexer.Store(ctx, docs)
	if err != nil {
		panic(err)
	}

	//var query string
	//for {
	//	_, _ = fmt.Scan(&query)
	//	output, err := r.Generate(ctx, query)
	//	if err != nil {
	//		panic(err)
	//	}
	//	var fullContent string // 用来拼接所有片段
	//	for {
	//		o, err := output.Recv()
	//		if err != nil {
	//			if err == io.EOF {
	//				break
	//			}
	//			panic(err) // 其他错误才 panic
	//		}
	//		if o.Content != "" {
	//			fullContent += o.Content
	//			fmt.Print(o.Content)
	//		}
	//	}
	//}
}
