package jjf

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func Buildmytest(ctx context.Context) (r compose.Runnable[map[string]interface{}, *schema.Message], err error) {
	const (
		ToolsNode1 = "ToolsNode1"
		ChatModel1 = "ChatModel1"
		Lambda3    = "Lambda3"
		Lambda4    = "Lambda4"
	)
	g := compose.NewGraph[map[string]interface{}, *schema.Message]()
	toolsNode1KeyOfToolsNode, err := newToolsNode(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddToolsNode(ToolsNode1, toolsNode1KeyOfToolsNode, compose.WithNodeName("加减法"),
		compose.WithInputKey("numbe1"), // 对应工具的number1参数
		compose.WithInputKey("numbe2")) // 对应工具的输出结果

	chatModel1KeyOfChatModel, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(ChatModel1, chatModel1KeyOfChatModel)
	//_ = g.AddLambdaNode(Lambda3, compose.InvokableLambda(newLambda))
	_ = g.AddLambdaNode(Lambda4, compose.InvokableLambda(ToolsResultToMessages))
	_ = g.AddEdge(compose.START, ToolsNode1)
	_ = g.AddEdge(ToolsNode1, Lambda4)
	_ = g.AddEdge(Lambda4, ChatModel1)
	_ = g.AddEdge(ChatModel1, compose.END)
	r, err = g.Compile(ctx, compose.WithGraphName("mytest"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
