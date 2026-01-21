package main

import (
	"context"
	"fmt"
	"log"
	"main/database"
	"main/jjf4"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/segmentio/kafka-go"
	"github.com/xuri/excelize/v2"
)

const (
	prefix = "OuterCyrex:"
	index  = "OuterIndex"
)

func jjf4sss(ctx context.Context) {
	database.InitRedis(ctx)
	database.InitMysql(ctx)

	r, err := jjf4.Buildmytest2(ctx)
	if err != nil {
		fmt.Printf("编译Graph流程失败：%v\n", err)
		return
	}
	maps := []*schema.Message{{
		Role:    schema.User,
		Content: "我要导出25年10月13号的用户玩法明细",
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

func kafkas() {
	topic := "my-topic"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "127.0.0.1:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: []byte("SELECT user_player.id, user_player.uid, user_player.gift_id, user_player.gift_num, user_player.gift_name, user_player.gift_image, user_player.player_id, user_player.player_detail_id, user_player.create_at, user_player.update_at, `user`.nickname FROM user_player JOIN `user` ON user_player.uid = `user`.id")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}

func re() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "my-topic",
		Partition: 0,
		MaxBytes:  10e6, // 10MB
		GroupID:   "my-first-kafka-consumer-group",
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}

func main() {
	//ctx := context.Background()
	//jjf4sss(ctx)
	ss := strconv.Itoa(int(time.Now().Unix()))

	fmt.Println(ss)
	//kafkas()
	//ttttt()
	//re()
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

func ttttt() {

	start := time.Now()
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Sheet1"
	sw, err := f.NewStreamWriter(sheet)
	must(err)

	// 写表头
	header := []interface{}{"ID", "Name", "Amount", "CreatedAt"}
	cell, _ := excelize.CoordinatesToCellName(1, 1)
	must(sw.SetRow(cell, header))

	// 模拟写 1,000,000 行
	n := 1_000_000
	for i := 1; i <= n; i++ {
		row := []interface{}{
			i,
			"User_" + strconv.Itoa(i),
			rand.Intn(10_000),
			time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339),
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		must(sw.SetRow(cell, row))

		// 每 50k 行打印一次内存占用
		if i%50_000 == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Printf("[progress] rows=%d heap=%.2fMB\n", i, float64(m.HeapAlloc)/1024.0/1024.0)
		}
	}

	must(sw.Flush())

	// 可选：冻结首行 + 自动列宽（注意：自动列宽对流式无感，需在 Flush 后做固定宽度）
	// 冻结首行
	must(f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		ActivePane:  "bottomLeft",
		TopLeftCell: "A2",
	}))

	// 保存到磁盘（也可改为写到 HTTP ResponseWriter，见后文）
	must(f.SaveAs("bigdata.xlsx"))

	fmt.Printf("done in %v\n", time.Since(start))

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
