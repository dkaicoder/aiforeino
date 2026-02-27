package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"main/graph/export_graph"
	"main/internal/model"
	"main/internal/repository"
	"main/pkg/ai"
	"main/pkg/common"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	ChatHistoryRepo repository.ChatHistoryRepository
	ExportGraph     *export_graph.ExportGraph
}

func NewAgent(chatHistoryRepo repository.ChatHistoryRepository, exportGraph *export_graph.ExportGraph) *Agent {
	return &Agent{
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
func (e *Agent) bigChatModel(ctx context.Context, question string, w http.ResponseWriter, flusher http.Flusher) *schema.StreamReader[*schema.Message] {
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
		&ExportTool{
			ExportGraph: e.ExportGraph,
		},
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
	streamResult, err := agent.Stream(execCtx, getChatHistoryFunc)
	return streamResult
}

func (e *Agent) StreamHandler(w http.ResponseWriter, r *http.Request) {
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
	chatMessage := &model.ChatMessage{
		MsgID:     messageId,
		Role:      schema.Assistant,
		Content:   reposeAnswer,
		Timestamp: time.Now().Unix(),
	}
	err := e.ChatHistoryRepo.SaveChatMessage(ctx, "session_12345", chatMessage)
	if err != nil {
		log.Fatalf("保存消息失败：%v", err)
	}
}

func (e *Agent) GetHis(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	getChatHistory, err := e.ChatHistoryRepo.GetChatHistory(ctx, "session_12345", 10, 0)
	if err != nil {
		panic(fmt.Errorf("读取历史对话失败：%w", err))
	}
	j, err := json.Marshal(getChatHistory)
	if err != nil {
		panic(err)
	}
	w.Write(j)
}
