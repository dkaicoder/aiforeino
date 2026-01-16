package jjf2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

// newLambda component initialization function of node 'Lambda3' in graph 'mytest'
func newLambda(ctx context.Context, input *schema.Message) (output *schema.Message, err error) {
	fmt.Println(input)
	arg := struct {
		Number  int `json:"number"`
		Number2 int `json:"number2"`
	}{}
	_ = json.Unmarshal([]byte(input.Content), &arg)
	fmts := fmt.Sprintf("%d", arg.Number+arg.Number2)
	return &schema.Message{
		Role:    schema.System,
		Content: fmts,
	}, nil
}
