package jjf4

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func Buildmytest2(ctx context.Context) (r compose.Runnable[[]*schema.Message, *schema.Message], err error) {
	const (
		ToolsNode1             = "ToolsNode1"
		TransformForEnd        = "TransformForEnd"
		ChatModel5             = "ChatModel5"
		Retriever1             = "Retriever1"
		TransformForRetriever  = "TransformForRetriever"
		ChatModel6             = "ChatModel6"
		TransformForModel      = "TransformForModel"
		TransformForFirstModel = "TransformForFirstModel"
	)
	g := compose.NewGraph[[]*schema.Message, *schema.Message](compose.WithGenLocalState(func(ctx context.Context) (state *MyGraphState) {
		return &MyGraphState{
			Query: "",
		}
	}))
	//toolToOutput := func(ctx context.Context, input string) ([]*schema.Message, error) {
	//	s := []*schema.Message{
	//		{
	//			Role:    schema.Assistant,
	//			Content: input,
	//		},
	//	}
	//	return s, nil
	//}

	toolsNode1KeyOfToolsNode, err := newToolsNode(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddToolsNode(ToolsNode1, toolsNode1KeyOfToolsNode)
	_ = g.AddLambdaNode(TransformForEnd, compose.InvokableLambda(newLambda))
	chatModel5KeyOfChatModel, err := newChatModelDoubao15pro(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(ChatModel5, chatModel5KeyOfChatModel)
	retriever1KeyOfRetriever, err := newRetriever(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddRetrieverNode(Retriever1, retriever1KeyOfRetriever)
	_ = g.AddLambdaNode(TransformForRetriever, compose.InvokableLambda(newLambda1))
	chatModel6KeyOfChatModel, err := newChatModel(ctx)

	toolImpl := &ToolImpl{}
	toolInfo, err := toolImpl.Info(context.Background())
	if err != nil {
		return nil, err
	}
	toolInfoList := []*schema.ToolInfo{toolInfo}
	if err := chatModel6KeyOfChatModel.BindTools(toolInfoList); err != nil {
		return nil, err
	}

	l2StateToOutput := func(ctx context.Context, input []*schema.Document, state *MyGraphState) ([]*schema.Document, error) {
		return input, nil
	}
	_ = g.AddChatModelNode(ChatModel6, chatModel6KeyOfChatModel)
	_ = g.AddLambdaNode(TransformForModel, compose.InvokableLambda(newLambda2), compose.WithStatePreHandler(l2StateToOutput))
	_ = g.AddLambdaNode(TransformForFirstModel, compose.InvokableLambda(newLambda3))
	_ = g.AddEdge(compose.START, TransformForFirstModel)
	_ = g.AddEdge(TransformForEnd, compose.END)
	_ = g.AddEdge(ChatModel6, ToolsNode1)
	_ = g.AddEdge(ToolsNode1, TransformForEnd)
	_ = g.AddEdge(TransformForFirstModel, ChatModel5)
	_ = g.AddEdge(ChatModel5, TransformForRetriever)
	_ = g.AddEdge(TransformForRetriever, Retriever1)
	_ = g.AddEdge(Retriever1, TransformForModel)
	_ = g.AddEdge(TransformForModel, ChatModel6)
	r, err = g.Compile(ctx, compose.WithGraphName("mytest2"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
