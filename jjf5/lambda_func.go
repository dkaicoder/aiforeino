package jjf5

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// newLambda component initialization function of node 'TransformForEnd' in graph 'mytest2'
func newLambda(ctx context.Context, input []*schema.Message) (output *schema.Message, err error) {
	panic("implement me")
}

// newLambda1 component initialization function of node 'TransformForRetriever' in graph 'mytest2'
func newLambda1(ctx context.Context, input *schema.Message) (output string, err error) {
	panic("implement me")
}

// newLambda2 component initialization function of node 'TransformForModel' in graph 'mytest2'
func newLambda2(ctx context.Context, input []*schema.Document) (output []*schema.Message, err error) {
	panic("implement me")
}

// newLambda3 component initialization function of node 'TransformForFirstModel' in graph 'mytest2'
func newLambda3(ctx context.Context, input string) (output []*schema.Message, err error) {
	panic("implement me")
}
