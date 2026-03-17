package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"main/config"
	"main/graph/export_graph"
	"main/pkg/common"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type ExportTool struct {
	ExportGraph *export_graph.ExportGraph
	C           *config.ParamsConfig
}

type ToolResult struct {
	Status string `json:"status"`
	TaskID string `json:"task_id"`
	Msg    string `json:"msg"`
}

func (e *ExportTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "数据导出工具",
		Desc: "当用户需要导出业务数据为Excel文件时使用。触发场景包括：用户提到订单、玩法、用户、导出、下载、Excel等关键词，或明确提出导出需求。示例：“帮我导出这个月的订单列表”、“我要下载用户数据Excel”、“把所有玩法配置导出来”。该工具将数据库表数据导出为Excel文件。" +
			"目前仅能导出玩法、账单、和用户资料数据，要求导出其他直接回复暂不支持",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"desc": {
				Desc:     "用户的原始导出需求描述，用于向量数据库检索匹配的业务表与导出规则。",
				Type:     schema.String,
				Required: false,
			},
			"graph_type": {
				Desc:     "图谱类型，固定值“export”。",
				Type:     schema.String,
				Required: false,
			},
		}),
	}, nil
}

func (e *ExportTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	msgChan := make(chan string, 10)
	exportTaskID := fmt.Sprintf("export_%d", time.Now().Unix())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range msgChan {
			if emitter, ok := common.GetProgressEmitter(ctx); ok {
				emitter.Emit(msg)
			}
		}
	}()

	_, err := e.RunExportGraph(ctx, exportTaskID, msgChan, argumentsInJSON)
	res := &ToolResult{}
	res.TaskID = exportTaskID
	if err != nil {
		res.Msg = err.Error()
		res.Status = "error"
	} else {
		path := fmt.Sprintf("导出任务已成功完成，文件可通过以下链接下载：%s/%s.xlsx", e.C.ExportHost, exportTaskID)
		res.Msg = path
		res.Status = "completed"
	}
	wg.Wait()
	jsonBytes, _ := json.Marshal(res)
	return string(jsonBytes), nil
}

func (e *ExportTool) RunExportGraph(ctx context.Context, exportTaskID string, msgChan chan string, questing string) ([]*schema.Message, error) {
	cbh, flusher := langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      e.C.Langfuse.Host,
		PublicKey: e.C.Langfuse.PublicKey,
		SecretKey: e.C.Langfuse.SecretKey,
	})
	callbacks.AppendGlobalHandlers(cbh)
	r, err := e.ExportGraph.Buildmytest2(ctx)
	if err != nil {
		fmt.Printf("编译Graph流程失败：%v\n", err)
		return nil, err
	}
	defer close(msgChan)
	messageBody := export_graph.GraphChoice{}
	json.Unmarshal([]byte(questing), &messageBody)
	messageBody.ExportTaskID = exportTaskID
	questings, _ := json.Marshal(messageBody)
	maps := []*schema.Message{{
		Role:    schema.User,
		Content: string(questings),
	}}
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if info.Name != "" {
				sseMsg := e.buildSSEEvent("startprogress", exportTaskID, info.Name, "start")
				msgChan <- sseMsg
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != "" {
				sseMsg := e.buildSSEEvent("endprogress", exportTaskID, info.Name, "end")
				msgChan <- sseMsg
			}
			return ctx
		}).
		Build()
	ree, err := r.Invoke(ctx, maps, compose.WithCallbacks(handler))
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s.xlsx", e.C.ExportHost, exportTaskID)
	sseMsg := fmt.Sprintf("event: progress\ndata: {\"task_id\":\"%s\",\"status\":\"completed\",\"url\":\"%s\"}\n\n", exportTaskID, url)
	//	e.graphEndSaveRes(ctx, url)
	msgChan <- sseMsg
	flusher()
	return ree, nil
}

func (e *ExportTool) buildSSEEvent(eventType string, taskID string, node string, status string) string {
	progress := map[string]any{
		"task_id": taskID,
		"node":    node,
		"status":  status,
		"time":    time.Now().Format("15:04:05"),
	}
	data, _ := json.Marshal(progress)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(data))
}
