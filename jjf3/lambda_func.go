package jjf3

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// newLambda component initialization function of node 'Lambda2' in graph 'mytest2'
func newLambda(ctx context.Context, input []*schema.Message) (output *schema.Message, err error) {
	return input[0], nil
}
