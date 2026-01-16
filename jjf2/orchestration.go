package jjf2

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func Buildmytest(ctx context.Context) (r compose.Runnable[[]*schema.Message, *schema.Message], err error) {
	const (
		ChatModel1 = "ChatModel1"
		Lambda3    = "Lambda3"
	)
	g := compose.NewGraph[[]*schema.Message, *schema.Message]()
	chatModel1KeyOfChatModel, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(ChatModel1, chatModel1KeyOfChatModel)
	_ = g.AddLambdaNode(Lambda3, compose.InvokableLambda(newLambda))
	_ = g.AddEdge(compose.START, ChatModel1)
	_ = g.AddEdge(Lambda3, compose.END)
	_ = g.AddEdge(ChatModel1, Lambda3)
	r, err = g.Compile(ctx, compose.WithGraphName("mytest"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
