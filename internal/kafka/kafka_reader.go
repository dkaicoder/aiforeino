package kafka

import (
	"context"
	"log"
	"main/internal/database"
	exports "main/internal/service/export"
	"sync"
)

type exportTask struct {
	sql      string
	exportId string
}

var ExportChan = make(chan exportTask, 100)
var ExportWg sync.WaitGroup

func Consumer(ctx context.Context) {
	reader := database.InitKafkaForConsumer(ctx)
	defer func() {
		// 消费协程退出时，关闭reader
		if err := reader.Close(); err != nil {
			log.Fatal("关闭Kafka reader失败:", err)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka消费协程：收到退出信号，停止消费")
			return
		default:
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Kafka读消息失败: %v，重试中...", err)
				continue
			}
			log.Printf("收到Kafka消息: offset=%d, value=%s", m.Offset, string(m.Value))
			exportSt := exportTask{
				sql:      string(m.Value),
				exportId: string(m.Key),
			}
			select {
			case ExportChan <- exportSt:
			case <-ctx.Done():
				log.Println("往通道塞数据时收到退出信号，放弃发送")
				return
			default:
				log.Printf("暂时无法处理任务: %v", m.Offset)
			}
		}
	}
}

// StartExportWorkers 启动协程池
func StartExportWorkers(ctx context.Context, workerCount int, export *exports.ExportService) {
	ExportWg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(workerID int, export *exports.ExportService) {
			defer ExportWg.Done()
			log.Printf("导出协程%d：启动", workerID)
			for {
				select {
				case <-ctx.Done():
					log.Printf("导出协程%d：收到退出信号，准备退出", workerID)
					return
				case data, ok := <-ExportChan:
					if !ok {
						log.Printf("导出协程%d：通道已关闭，无数据可处理，退出", workerID)
						return
					}
					if err := export.ExportData(ctx, data.sql, data.exportId); err != nil {
						log.Printf("导出协程%d：数据[%s]导出失败: %v", workerID, data, err)
					} else {
						log.Printf("导出协程%d：数据[%s]导出成功", workerID, data)
					}
				}
			}
		}(i, export)
	}
}
