package jjf

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// newLambda component initialization function of node 'Lambda3' in graph 'mytest'
func newLambda(ctx context.Context, input map[string]string) (output any, err error) {
	return input, nil
}

func ToolsResultToMessages(ctx context.Context, input map[string]interface{}) ([]*schema.Message, error) {
	// 直接将工具返回的纯数字字符串封装为 schema.Message
	return []*schema.Message{
		{
			Role:    schema.System,
			Content: "123123123",
		},
	}, nil
}
