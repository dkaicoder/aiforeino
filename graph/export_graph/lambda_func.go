package export_graph

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/database"
	"main/internal/model"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/segmentio/kafka-go"
)

// newLambda component initialization function of node 'TransformForEnd' in graph 'mytest2'
func newLambda(ctx context.Context, input []*schema.Message) (output []*schema.Message, err error) {
	res := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   string `json:"data"`
	}{}
	_ = json.Unmarshal([]byte(input[0].Content), &res)
	if res.Data == "" {
		return nil, fmt.Errorf(res.Msg)
	}
	checkSQL := fmt.Sprintf("SELECT 1 FROM (%s) AS t LIMIT 1", res.Data)
	var exists int
	err = database.MysqlDb.Raw(checkSQL).Scan(&exists).Error
	if err != nil {
		return nil, fmt.Errorf("检查数据存在性失败: %v", err)
	}
	if exists != 1 {
		return nil, fmt.Errorf("改条件范围没有任何数据，请调整后重新导出")
	}
	var state *MyGraphState
	err = compose.ProcessState[*MyGraphState](ctx, func(ctx context.Context, s *MyGraphState) error {
		state = s
		return nil
	})
	if err != nil {
		return nil, err
	}
	exportId := state.ExportTaskID

	go func() {
		downloadList := &model.DownloadList{
			Name:       exportId,
			CreateTime: time.Now(),
			Type:       "xlsx",
		}
		state.DownloadRepo.CrateTask(downloadList)
	}()

	kafkaConn := database.InitKafkaForProducer(ctx)
	_, err = kafkaConn.WriteMessages(
		kafka.Message{Value: []byte(res.Data), Key: []byte(exportId)},
	)
	if err != nil {
		return nil, err
	}
	return input, nil
}

// newLambda1 component initialization function of node 'TransformForRetriever' in graph 'mytest2'
func newLambda1(ctx context.Context, input *schema.Message) (output string, err error) {
	return input.Content, nil
}

// newLambda2 component initialization function of node 'TransformForModel' in graph 'mytest2'
func newLambda2(ctx context.Context, input []*schema.Document) (output []*schema.Message, err error) {

	var state *MyGraphState
	err = compose.ProcessState[*MyGraphState](ctx, func(ctx context.Context, s *MyGraphState) error {
		state = s
		return nil
	})
	if err != nil {
		return nil, err
	}
	query := state.Query

	var docsContent string
	for _, doc := range input {
		docsContent += doc.Content + "\n"
	}
	tpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(`问题: {content}`))
	messages, err := tpl.Format(ctx, map[string]any{
		"documents": docsContent,
		"content":   query,
	})
	return messages, err
}

// newLambda3 component initialization function of node 'TransformForFirstModel' in graph 'mytest2'
func newLambda3(ctx context.Context, input []*schema.Message) (output []*schema.Message, err error) {
	graphChoice := GraphChoice{}
	json.Unmarshal([]byte(input[0].Content), &graphChoice)
	question := graphChoice.Desc
	_ = compose.ProcessState[*MyGraphState](ctx, func(ctx context.Context, state *MyGraphState) error {
		state.Query = question
		state.ExportTaskID = graphChoice.ExportTaskID
		return nil
	})

	systemRole := &schema.Message{
		Role: schema.System,
		Content: "你的任务是：" +
			"1.从用户的自然语言请求中，识别出「业务名词对应的数据库表名」（只返回表名，用顿号分隔）；" +
			"2.忽略时间、操作（如导出/查询/给我）等无关内容；\n3. 若涉及关联查询，需识别出所有相关表名。" +
			"示例：" +
			"1.用户说“给我下昨天xx商品的订单明细” → 返回“商品表、订单表”；" +
			"2. 用户说“导出今天的订单数据” → 返回“订单表”；" +
			"3. 用户说“查用户的商品购买记录” → 返回“用户表、商品表、订单表”；" +
			"4. 用户说“昨天的销售额统计” → 返回“订单表、商品表”。",
	}
	userRole := &schema.Message{
		Role:    schema.User,
		Content: question,
	}
	return []*schema.Message{systemRole, userRole}, nil
}

//func newLambdaForNeed(ctx context.Context, input []*schema.Message) (output []*schema.Message, err error) {
//	question := input[0].Content
//	messageId := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
//	chatMessage := &ChatMessage{
//		MsgID:     messageId,
//		Role:      schema.User,
//		Content:   question,
//		Timestamp: time.Now().Unix(),
//	}
//
//	getChatHistory, err := chatMessage.GetChatHistory(ctx, "session_12345", 10, 0)
//	if err != nil {
//		return nil, fmt.Errorf("读取历史对话失败：%w", err)
//	}
//	for _, msg := range getChatHistory {
//		historyMsg := &schema.Message{
//			Role:    msg.Role,
//			Content: msg.Content,
//		}
//		output = append(output, historyMsg)
//	}
//
//	err = chatMessage.SaveChatMessage(ctx, "session_12345")
//	if err != nil {
//		return nil, fmt.Errorf("保存消息失败：%w", err)
//	}
//
//	var toolList = map[string]string{
//		"export": "导出数据库表数据到excel文件",
//	}
//	toolDesc := "支持的工具列表：\n"
//	for t, desc := range toolList {
//		toolDesc += fmt.Sprintf("- %s：%s\n", t, desc)
//	}
//	tpl := prompt.FromMessages(schema.FString,
//		schema.SystemMessage(userPrompt),
//		schema.UserMessage(question))
//	messages, err := tpl.Format(ctx, map[string]any{
//		"tool_list": toolDesc,
//	})
//	output = append(output, messages...)
//	return output, nil
//}

func newLambdaForArr(ctx context.Context, input []*schema.Message) (output []*schema.Message, err error) {
	return input, nil
}
