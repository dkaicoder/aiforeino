package export_graph

import (
	"context"
	"main/pkg/llm"

	"github.com/cloudwego/eino/components/model"
)

// newChatModel component initialization function of node 'ChatModel5' in graph 'mytest2'
func newChatModel(ctx context.Context) (cm model.ChatModel, err error) {
	cm, err = llm.NewChatModelFactory(ctx, "doubao-1-5-pro-32k-250115")
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func newChatModelDoubao15pro(ctx context.Context) (cm model.ChatModel, err error) {
	cm, err = llm.NewChatModelFactory(ctx, "doubao-1-5-pro-32k-250115")
	if err != nil {
		return nil, err
	}
	return cm, nil
}
