package export

import (
	"context"
	"encoding/json"
	"fmt"
	"main/config"
	"main/graph/export_graph"
	"main/pkg/progress"
	"time"

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
	exportTaskID := fmt.Sprintf("export_%d", time.Now().Unix())
	_, err := e.RunExportGraph(ctx, exportTaskID, argumentsInJSON)
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
	jsonBytes, _ := json.Marshal(res)
	return string(jsonBytes), nil
}

func (e *ExportTool) RunExportGraph(ctx context.Context, exportTaskID string, questing string) ([]*schema.Message, error) {
	r, err := e.ExportGraph.Buildmytest2(ctx)
	if err != nil {
		fmt.Printf("编译Graph流程失败：%v\n", err)
		return nil, err
	}
	messageBody := export_graph.GraphChoice{}
	json.Unmarshal([]byte(questing), &messageBody)
	messageBody.ExportTaskID = exportTaskID
	questings, _ := json.Marshal(messageBody)
	maps := []*schema.Message{{
		Role:    schema.User,
		Content: string(questings),
	}}
	now := func() string { return time.Now().Format("15:04:05") }
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if info.Name != "" {
				progress.TryPublish(ctx, progress.ProgressEvent{
					Kind:   progress.KindStepStart,
					TaskID: exportTaskID,
					Node:   info.Name,
					Time:   now(),
				})
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != "" {
				progress.TryPublish(ctx, progress.ProgressEvent{
					Kind:   progress.KindStepEnd,
					TaskID: exportTaskID,
					Node:   info.Name,
					Time:   now(),
				})
			}
			return ctx
		}).
		Build()
	ree, err := r.Invoke(ctx, maps, compose.WithCallbacks(handler))
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s.xlsx", e.C.ExportHost, exportTaskID)
	progress.TryPublish(ctx, progress.ProgressEvent{
		Kind:         progress.KindExportComplete,
		TaskID:       exportTaskID,
		ExportStatus: "completed",
		URL:          url,
	})
	return ree, nil
}
