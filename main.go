package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"main/config"
	"main/database"
	"main/exportAi"
	"main/pkg/ai"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	uuid2 "github.com/google/uuid"
)

const (
	prefix = "OuterCyrex:"
	index  = "OuterIndex"
)

type ProgressEmitter interface {
	Emit(msg string)
}
type progressKeyType struct{}

var progressKey = progressKeyType{}

func WithProgressEmitter(
	ctx context.Context,
	emitter ProgressEmitter,
) context.Context {
	return context.WithValue(ctx, progressKey, emitter)
}

func GetProgressEmitter(ctx context.Context) (ProgressEmitter, bool) {
	emitter, ok := ctx.Value(progressKey).(ProgressEmitter)
	return emitter, ok
}

type LogEmitter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (l *LogEmitter) Emit(msg string) {
	fmt.Fprint(l.w, msg)
	l.flusher.Flush()
}

type exportTool struct{}

func (e *exportTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "数据导出工具",
		Desc: "当用户需要导出业务数据为Excel文件时使用。触发场景包括：用户提到订单、玩法、用户、导出、下载、Excel等关键词，或明确提出导出需求。示例：“帮我导出这个月的订单列表”、“我要下载用户数据Excel”、“把所有玩法配置导出来”。该工具将数据库表数据导出为Excel文件。",
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
			"export_task_id": {
				Desc:     "随机生成的导出任务ID，用于标识本次导出请求。",
				Type:     schema.String,
				Required: false,
			},
		}),
	}, nil
}

func (e *exportTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	msgChan := make(chan string, 100)
	exportTaskID := fmt.Sprintf("export_%d", time.Now().Unix())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range msgChan {
			if emitter, ok := GetProgressEmitter(ctx); ok {
				emitter.Emit(msg)
			}
		}
	}()
	go func(exportTaskID string) {
		defer close(msgChan)
		jjf4sss(ctx, exportTaskID, msgChan, argumentsInJSON)
	}(exportTaskID)
	wg.Wait()
	return `{"code":200,"msg":"导出成功"}`, nil
}

func jjf4sss(ctx context.Context, exportTaskID string, msgChan chan string, questing string) []*schema.Message {
	cbh, flusher := langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      "https://us.cloud.langfuse.com",
		PublicKey: "pk-lf-8f9c06d7-b5ff-48ed-a559-e0e04d197e88",
		SecretKey: "sk-lf-9d1c1bf0-6458-41a2-aa40-9a5b56015c06",
	})
	callbacks.AppendGlobalHandlers(cbh)
	r, err := exportAi.Buildmytest2(ctx)
	if err != nil {
		fmt.Printf("编译Graph流程失败：%v\n", err)
		return nil
	}

	messageBody := exportAi.GraphChoice{}
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
				sseMsg := buildSSEEvent("startprogress", exportTaskID, info.Name, "start")
				msgChan <- sseMsg
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Name != "" {
				sseMsg := buildSSEEvent("endprogress", exportTaskID, info.Name, "end")
				msgChan <- sseMsg
			}
			return ctx
		}).
		Build()
	ree, err := r.Invoke(ctx, maps, compose.WithCallbacks(handler))
	if err != nil {
		msg := strings.ReplaceAll(err.Error(), "\n", " | ")
		sseMsg := fmt.Sprintf(
			"event: progress\ndata: {\"task_id\":\"%s\",\"status\":\"error\",\"msg\":\"%s\"}\n\n",
			exportTaskID,
			msg,
		)
		graphEndSaveRes(ctx, err.Error())
		msgChan <- sseMsg
	} else {
		url := "http://127.0.0.1:8080/" + exportTaskID + ".xlsx"
		sseMsg := fmt.Sprintf("event: progress\ndata: {\"task_id\":\"%s\",\"status\":\"completed\",\"url\":\"%s\"}\n\n", exportTaskID, url)
		graphEndSaveRes(ctx, url)
		msgChan <- sseMsg
	}
	flusher()
	return ree
}

func graphEndSaveRes(ctx context.Context, content string) (err error) {
	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
	chatMessage := &ChatMessage{
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

func getChatHistory(ctx context.Context, question string) (output []*schema.Message, err error) {
	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
	chatMessage := &ChatMessage{
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

func buildSSEEvent(eventType string, taskID string, node string, status string) string {
	progress := map[string]any{
		"task_id": taskID,
		"node":    node,
		"status":  status,
		"time":    time.Now().Format("15:04:05"),
	}
	data, _ := json.Marshal(progress)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(data))
}

// 前置需求验证
func frontChatModel(ctx context.Context, question string) string {
	chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: "358ad2c2-9d3b-4990-92c7-117cf25fdae3",
		Model:  "doubao-seed-1-6-lite-251015",
	})
	var output []*schema.Message
	output = append(output, &schema.Message{
		Role:    schema.System,
		Content: roleForFrontModel(),
	})
	output = append(output, &schema.Message{
		Role:    schema.User,
		Content: question,
	})
	generateResult, err := chatModel.Generate(ctx, output)
	if err != nil || generateResult == nil {
		log.Fatalf("创建流失败: %v", err)
	}
	return generateResult.Content
}

// 大模型聊天
func bigChatModel(ctx context.Context, question string, w http.ResponseWriter, flusher http.Flusher) *schema.StreamReader[*schema.Message] {
	chatModel, err := ai.NewChatModelFactory(ctx, "doubao-seed-1-6-251015")
	if err != nil {
		panic(err)
	}
	emitter := &LogEmitter{
		w:       w,
		flusher: flusher,
	}
	execCtx := WithProgressEmitter(ctx, emitter)
	toolList := []tool.BaseTool{
		&exportTool{},
	}
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: toolList},
	})

	getChatHistoryFunc, err := getChatHistory(ctx, question)
	if err != nil {
		log.Fatalf("获取历史对话失败: %v", err)
	}
	userQ := &schema.Message{
		Role:    schema.User,
		Content: question,
	}
	system := &schema.Message{
		Role:    schema.System,
		Content: "你是一个精简回答助手，所有回答都要简洁明了，控制在100字以内，只输出核心结论，避免冗余解释。",
	}
	getChatHistoryFunc = append(getChatHistoryFunc, userQ, system)
	streamResult, err := agent.Stream(execCtx, getChatHistoryFunc)

	//getChatHistoryFunc, err := getChatHistory(ctx, question)
	//if err != nil {
	//	log.Fatalf("获取历史对话失败: %v", err)
	//}
	//template := prompt.FromMessages(schema.FString,
	//	schema.SystemMessage(roleForBigModel()),
	//	schema.MessagesPlaceholder("chat_history", true),
	//	schema.UserMessage("问题: {question}"),
	//)
	//var toolList = map[string]string{
	//	"export": "导出数据库表数据到excel文件",
	//}
	//toolDesc := "支持的工具列表：\n"
	//for t, desc := range toolList {
	//	toolDesc += fmt.Sprintf("- %s：%s\n", t, desc)
	//}
	//messages, err := template.Format(context.Background(), map[string]any{
	//	//"tool_list":    toolDesc,
	//	"question":     question,
	//	"chat_history": getChatHistoryFunc,
	//})
	//streamResult, err := chatModel.Stream(ctx, messages)
	//if err != nil || streamResult == nil {
	//	log.Fatalf("创建流失败: %v", err)
	//}
	return streamResult
}

func roleForBigModel() string {
	var userPrompt = `
	Role: 智能助手（兼工具识别与对话）
	核心设定：
	你是K的专属智能助手，负责处理用户提问并识别导出需求。
	- 当被问及身份时，回答：我是K的智能助手
	- 当被问及能做什么时，回答：目前我可以处理部分导出业务，以及解答日常问题
	
	对话规则：
	1. 自然语言对话
	   对于闲聊、问候、解释概念等非导出类提问，用友好、专业的自然语言回答
	   无法回答的问题（如实时天气），可引导用户使用导出功能，例如：我无法查询实时天气，但可以帮您处理导出业务或解答其他日常问题
	
	2. 格式要求
	   所有回答必须是纯自然语言，禁止返回任何JSON格式内容,尽量回答内容简介明了。
    `
	return userPrompt
}

func roleForFrontModel() string {
	var userPrompt = `
		# Role: 前置需求判断助手
		## 核心目标
		判断用户提问是否包含**导出/数据查询需求**，仅返回符合要求的 JSON 格式，无需求时返回空字符串。
		
		### 判定条件（必须同时满足）
		1.  用户提问包含关键词：导出、数据、明细、订单、玩法、报表、下载、查询、统计、分析
		2.  用户提问包含限定词：时间、用户、订单号、范围
		
		> 注：若缺少限定词，需返回「请补充需要导出的具体信息」
		
		### 输出格式要求
		- 要么必须返回完整 JSON，无任何前缀、解释或 markdown或者要么返回空字符串
		- 禁止返回自然语言
		- JSON 格式：
		{"graph_type":"export","desc":"一句话概括用户需求，用于后续查询"}
`
	return userPrompt
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
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

	stream := bigChatModel(ctx, question, w, flusher)
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
	chatMessage := &ChatMessage{
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

func getHis(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	chatMessage := &ChatMessage{}
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

func main() {
	ctx := context.Background()
	configs := config.InitConfig()
	database.Init(configs)
	database.InitRedis(ctx)
	database.InitMysql(ctx)
	fileServer := http.FileServer(http.Dir("static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/chat/history", getHis)
	http.HandleFunc("/stream", streamHandler)
	http.ListenAndServe(":8080", nil)
}

func save(ctx context.Context) {
	r, err := InitRAGEngine(ctx, index, prefix)
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
