package export

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	exportgraph2 "main/graph/export_graph"
	"main/internal/repository"
	"main/pkg/ai"
	"main/pkg/common"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type ExportService struct {
}

func (e *ExportService) AiChatDo(ctx context.Context, exportTaskID string, msgChan chan string, questing string) ([]*schema.Message, error) {
	cbh, flusher := langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      "https://us.cloud.langfuse.com",
		PublicKey: "pk-lf-8f9c06d7-b5ff-48ed-a559-e0e04d197e88",
		SecretKey: "sk-lf-9d1c1bf0-6458-41a2-aa40-9a5b56015c06",
	})
	callbacks.AppendGlobalHandlers(cbh)
	r, err := exportgraph2.Buildmytest2(ctx)
	if err != nil {
		fmt.Printf("编译Graph流程失败：%v\n", err)
		return nil, err
	}
	defer close(msgChan)
	messageBody := exportgraph2.GraphChoice{}
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
	url := "http://192.168.3.182:8080/" + exportTaskID + ".xlsx"
	sseMsg := fmt.Sprintf("event: progress\ndata: {\"task_id\":\"%s\",\"status\":\"completed\",\"url\":\"%s\"}\n\n", exportTaskID, url)
	//	e.graphEndSaveRes(ctx, url)
	msgChan <- sseMsg
	flusher()
	return ree, nil
}

func (e *ExportService) graphEndSaveRes(ctx context.Context, content string) (err error) {
	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
	chatMessage := &repository.ChatMessage{
		MsgID:     messageId,
		Role:      schema.Assistant,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}
	err = chatMessage.SaveChatMessage(ctx, "session_12345")
	if err != nil {
		return fmt.Errorf("保存消息失败：%w", err)
	}
	return nil
}

func (e *ExportService) getChatHistory(ctx context.Context, question string) (output []*schema.Message, err error) {
	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
	chatMessage := &repository.ChatMessage{
		MsgID:     messageId,
		Role:      schema.User,
		Content:   question,
		Timestamp: time.Now().Unix(),
	}

	getChatHistory, err := chatMessage.GetChatHistory(ctx, "session_12345", 10, 0)
	if err != nil {
		return nil, fmt.Errorf("读取历史对话失败：%w", err)
	}
	for _, msg := range getChatHistory {
		historyMsg := &schema.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
		output = append(output, historyMsg)
	}

	err = chatMessage.SaveChatMessage(ctx, "session_12345")
	if err != nil {
		return nil, fmt.Errorf("保存消息失败：%w", err)
	}
	return output, nil
}

func (e *ExportService) buildSSEEvent(eventType string, taskID string, node string, status string) string {
	progress := map[string]any{
		"task_id": taskID,
		"node":    node,
		"status":  status,
		"time":    time.Now().Format("15:04:05"),
	}
	data, _ := json.Marshal(progress)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(data))
}

// 大模型聊天
func (e *ExportService) bigChatModel(ctx context.Context, question string, w http.ResponseWriter, flusher http.Flusher) *schema.StreamReader[*schema.Message] {
	chatModel, err := ai.NewChatModelFactory(ctx, "doubao-1-5-pro-32k-250115")
	if err != nil {
		panic(err)
	}
	emitter := &common.LogEmitter{
		W:       w,
		Flusher: flusher,
	}
	execCtx := common.WithProgressEmitter(ctx, emitter)
	toolList := []tool.BaseTool{
		&ExportTool{},
	}
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: toolList},
	})

	getChatHistoryFunc, err := e.getChatHistory(ctx, question)
	if err != nil {
		log.Fatalf("获取历史对话失败: %v", err)
	}
	userQ := &schema.Message{
		Role:    schema.User,
		Content: question,
	}
	prompt := `
		1. 你是一个精简回答助手，所有回答都要简洁明了，控制在100字以内，只输出核心结论，避免冗余解释。
		2. 当你调用工具后，如果工具返回的结果中包含 "status": "completed"这意味着任务已经圆满完成并且立即停止调用任何工具，直接基于工具返回的结果整理成自然语言回复。
		3. 当你调用工具后，如果工具返回的结果中包含 "status": "error"这意味着任务执行过程中出现了问题，基于工具返回的错误信息进行简要分析并回复用户。
	  `

	system := &schema.Message{
		Role:    schema.System,
		Content: prompt,
	}
	getChatHistoryFunc = append(getChatHistoryFunc, userQ, system)
	streamResult, err := agent.Stream(execCtx, getChatHistoryFunc)
	return streamResult
}

func (e *ExportService) StreamHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 设置响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	question := r.URL.Query().Get("question")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	stream := e.bigChatModel(ctx, question, w, flusher)
	defer stream.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var reposeAnswer string
	for {
		response, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Fprintf(w, "data: Error: %v\n\n", err)
			flusher.Flush()
			break
		}
		reposeAnswer += response.Content
		fmt.Fprintf(w, "data: %s\n\n", response.Content)
		flusher.Flush()
	}
	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
	chatMessage := &repository.ChatMessage{
		MsgID:     messageId,
		Role:      schema.Assistant,
		Content:   reposeAnswer,
		Timestamp: time.Now().Unix(),
	}
	err := chatMessage.SaveChatMessage(ctx, "session_12345")
	if err != nil {
		log.Fatalf("保存消息失败：%v", err)
	}
}

func (e *ExportService) GetHis(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	chatMessage := &repository.ChatMessage{}
	getChatHistory, err := chatMessage.GetChatHistory(ctx, "session_12345", 10, 0)
	if err != nil {
		panic(fmt.Errorf("读取历史对话失败：%w", err))
	}
	j, err := json.Marshal(getChatHistory)
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

type ExportTool struct{}

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

type ToolResult struct {
	Status string `json:"status"`
	TaskID string `json:"task_id"`
	Msg    string `json:"msg"`
}

func (e *ExportTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	exportService := &ExportService{}
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

	_, err := exportService.AiChatDo(ctx, exportTaskID, msgChan, argumentsInJSON)
	res := &ToolResult{}
	res.TaskID = exportTaskID
	if err != nil {
		res.Msg = err.Error()
		res.Status = "error"
	} else {
		res.Msg = "导出任务已成功完成，文件可通过以下链接下载：http://192.168.3.182:8080/" + exportTaskID + ".xlsx"
		res.Status = "completed"
	}
	wg.Wait()
	jsonBytes, _ := json.Marshal(res)
	return string(jsonBytes), nil
}
