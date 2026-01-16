package jjf3

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func Buildmytest2(ctx context.Context) (r compose.Runnable[[]*schema.Message, *schema.Message], err error) {
	const (
		ToolsNode1 = "ToolsNode1"
		Lambda2    = "Lambda2"
		ChatModel5 = "ChatModel5"
	)
	g := compose.NewGraph[[]*schema.Message, *schema.Message]()
	toolsNode1KeyOfToolsNode, err := newToolsNode(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddToolsNode(ToolsNode1, toolsNode1KeyOfToolsNode)
	_ = g.AddLambdaNode(Lambda2, compose.InvokableLambda(newLambda))
	chatModel5KeyOfChatModel, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	toolImpl := &ToolImpl{}
	toolInfo, err := toolImpl.Info(context.Background())
	if err != nil {
		return nil, err
	}
	toolInfoList := []*schema.ToolInfo{toolInfo}

	if err := chatModel5KeyOfChatModel.BindTools(toolInfoList); err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(ChatModel5, chatModel5KeyOfChatModel)
	_ = g.AddEdge(compose.START, ChatModel5)
	_ = g.AddEdge(Lambda2, compose.END)
	_ = g.AddEdge(ChatModel5, ToolsNode1)
	_ = g.AddEdge(ToolsNode1, Lambda2)
	r, err = g.Compile(ctx, compose.WithGraphName("mytest2"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
