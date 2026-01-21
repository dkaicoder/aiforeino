package jjf4

import (
	"context"
	"encoding/json"
	"fmt"
	"main/database"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/segmentio/kafka-go"
)

// newLambda component initialization function of node 'TransformForEnd' in graph 'mytest2'
func newLambda(ctx context.Context, input []*schema.Message) (output *schema.Message, err error) {
	res := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   string `json:"data"`
	}{}

	_ = json.Unmarshal([]byte(input[0].Content), &res)
	if res.Data == "" {
		return nil, fmt.Errorf(res.Msg)
	}
	kafkaConn := database.InitKafkaForProducer(ctx)
	_, err = kafkaConn.WriteMessages(
		kafka.Message{Value: []byte(res.Data)},
	)
	if err != nil {
		return nil, err
	}
	return input[0], nil
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
	_ = compose.ProcessState[*MyGraphState](ctx, func(ctx context.Context, state *MyGraphState) error {
		state.Query = input[0].Content
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
		Content: input[0].Content,
	}
	return []*schema.Message{systemRole, userRole}, nil
}
