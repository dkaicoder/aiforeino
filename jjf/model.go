package jjf

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

// newChatModel component initialization function of node 'ChatModel1' in graph 'mytest'
func newChatModel(ctx context.Context) (cm model.BaseChatModel, err error) {
	// TODO Modify component configuration here.
	config := &ark.ChatModelConfig{
		APIKey: "358ad2c2-9d3b-4990-92c7-117cf25fdae3",
		Model:  "doubao-seed-1-6-251015",
	}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
