package main

import (
	"context"
	"fmt"
	"main/jjf4"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	prefix = "OuterCyrex:"
	index  = "OuterIndex"
)

func jjf4sss(ctx context.Context) {
	jjf4.InitRedis(ctx)
	jjf4.InitMysql(ctx)

	r, err := jjf4.Buildmytest2(ctx)
	if err != nil {
		fmt.Printf("编译Graph流程失败：%v\n", err)
		return
	}
	maps := []*schema.Message{{
		Role:    schema.User,
		Content: "我要导出用户玩法明细",
	}}
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			fmt.Printf("[%s] >>> 节点开始: %s\n", time.Now().Format("15:04:05"), info.Name)
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			fmt.Printf("[%s] <<< 节点结束: %s\n", time.Now().Format("15:04:05"), info.Name)
			return ctx
		}).
		Build()
	result, err := r.Invoke(ctx, maps, compose.WithCallbacks(handler))
	if err != nil {
		fmt.Printf("运行流程失败：%v\n", err.Error())
		return
	}

	//
	fmt.Println(result)
}

func main() {
	ctx := context.Background()
	jjf4sss(ctx)

	////tool.Tool(ctx)
	//
	//r, err := InitRAGEngine(ctx, index, prefix)
	//if err != nil {
	//	panic(err)
	//}

	//doc, err := r.Loader.Load(ctx, document.Source{
	//	URI: "./test_txt/mysql-1.md",
	//})
	//if err != nil {
	//	panic(err)
	//}
	//
	//docs, err := r.Splitter.Transform(ctx, doc)
	//if err != nil {
	//	panic(err)
	//}
	//
	//for _, d := range docs {
	//	uuid, _ := uuid2.NewUUID()
	//	d.ID = uuid.String()
	//}
	//
	//err = r.InitVectorIndex(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//
	//_, err = r.Indexer.Store(ctx, docs)
	//if err != nil {
	//	panic(err)
	//}
	//
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
