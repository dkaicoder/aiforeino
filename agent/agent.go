package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	tool2 "main/agent/tool/export"
	"main/config"
	"main/graph/export_graph"
	"main/internal/model"
	"main/internal/observability"
	"main/internal/repository"
	"main/pkg/llm"
	"main/pkg/progress"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type Agent struct {
	C               *config.ParamsConfig
	ChatHistoryRepo repository.ChatHistoryRepository
	ExportGraph     *export_graph.ExportGraph
}

func NewAgent(c *config.ParamsConfig, chatHistoryRepo repository.ChatHistoryRepository, exportGraph *export_graph.ExportGraph) *Agent {
	return &Agent{
		C:               c,
		ChatHistoryRepo: chatHistoryRepo,
		ExportGraph:     exportGraph,
	}
}

func (e *Agent) graphEndSaveRes(ctx context.Context, content string) (err error) {
	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
	chatMessage := &model.ChatMessage{
		MsgID:     messageId,
		Role:      schema.Assistant,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}
	err = e.ChatHistoryRepo.SaveChatMessage(ctx, "session_12345", chatMessage)
	if err != nil {
		return fmt.Errorf("保存消息失败：%w", err)
	}
	return nil
}

func (e *Agent) getChatHistory(ctx context.Context, question string) (output []*schema.Message, err error) {
	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
	chatMessage := &model.ChatMessage{
		MsgID:     messageId,
		Role:      schema.User,
		Content:   question,
		Timestamp: time.Now().Unix(),
	}

	getChatHistory, err := e.ChatHistoryRepo.GetChatHistory(ctx, "session_12345", 10, 0)
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

	err = e.ChatHistoryRepo.SaveChatMessage(ctx, "session_12345", chatMessage)
	if err != nil {
		return nil, fmt.Errorf("保存消息失败：%w", err)
	}
	return output, nil
}

// 大模型聊天
func (e *Agent) bigChatModel(ctx context.Context, question string, w http.ResponseWriter, flusher http.Flusher) (*schema.StreamReader[*schema.Message], error) {
	chatModel, err := llm.NewChatModelFactory(ctx, e.C.ChatModel)
	if err != nil {
		return nil, fmt.Errorf("初始化大模型失败: %w", err)
	}
	toolList := []tool.BaseTool{
		&tool2.ExportTool{
			ExportGraph: e.ExportGraph,
			C:           e.C,
		},
	}
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: toolList},
	})

	if err != nil {
		return nil, fmt.Errorf("初始化智能体失败: %w", err)
	}

	getChatHistoryFunc, err := e.getChatHistory(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("获取历史对话失败: %w", err)
	}
	userQ := &schema.Message{
		Role:    schema.User,
		Content: question,
	}
	prompt := `
		1. 精简回答，100字内，只给核心结论。
		2. 工具调用铁律：
		   - 先读取工具描述中的核心操作（如导出、查询、修改等），仅当用户输入包含与工具核心操作匹配的明确动词，且指定操作对象（如数据类型）时，才调用对应工具。
		   - 意图模糊或仅提及操作对象关键词但无匹配动词时，先澄清或自然回应，不默认触发。
		3. 工具返回status:completed则停止调用并整理结果；status:error则简要说明问题。
		4. 绝对不调用工具：用户仅提问、吐槽、测试或询问模型/知识库信息，即使包含操作对象关键词，也只口头回应。
	  `

	system := &schema.Message{
		Role:    schema.System,
		Content: prompt,
	}
	getChatHistoryFunc = append(getChatHistoryFunc, userQ, system)
	streamResult, err := agent.Stream(ctx, getChatHistoryFunc)
	if err != nil {
		return nil, fmt.Errorf("发起流式响应失败: %w", err)
	}
	return streamResult, nil
}

func (e *Agent) StreamHandler(g *gin.Context) {
	defer observability.FlushLangfuse()

	w := g.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	question := g.Query("question")
	if question == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	evCh := make(chan progress.ProgressEvent, 64)
	sink := progress.NewChanSink(evCh)
	baseCtx := g.Request.Context()
	ctx := progress.WithSink(baseCtx, sink)

	var sseMu sync.Mutex
	withSSE := func(fn func()) {
		sseMu.Lock()
		defer sseMu.Unlock()
		fn()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for ev := range evCh {
			withSSE(func() {
				_ = progress.WriteSSE(w, flusher, ev)
			})
		}
	}()
	defer func() {
		close(evCh)
		<-done
	}()

	stream, err := e.bigChatModel(ctx, question, w, flusher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	saveCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var reposeAnswer string
	for {
		response, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			withSSE(func() {
				fmt.Fprintf(w, "data: Error: %v\n\n", err)
				flusher.Flush()
			})
			break
		}
		reposeAnswer += response.Content
		withSSE(func() {
			fmt.Fprintf(w, "data: %s\n\n", response.Content)
			flusher.Flush()
		})
	}

	withSSE(func() {
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	})

	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
	chatMessage := &model.ChatMessage{
		MsgID:     messageId,
		Role:      schema.Assistant,
		Content:   reposeAnswer,
		Timestamp: time.Now().Unix(),
	}
	err = e.ChatHistoryRepo.SaveChatMessage(saveCtx, "session_12345", chatMessage)
	if err != nil {
		log.Printf("保存消息失败: %v", err)
	}
}

func (e *Agent) GetHis(g *gin.Context) {
	ctx := context.Background()
	getChatHistory, err := e.ChatHistoryRepo.GetChatHistory(ctx, "session_12345", 10, 0)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取历史对话失败: %v", err)})
		return
	}
	g.JSON(http.StatusOK, getChatHistory)
}
