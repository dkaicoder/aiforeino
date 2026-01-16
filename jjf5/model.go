package jjf5

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

// newChatModel component initialization function of node 'ChatModel5' in graph 'mytest2'
func newChatModel(ctx context.Context) (cm model.ChatModel, err error) {
	// TODO Modify component configuration here.
	config := &ark.ChatModelConfig{
		APIKey: "358ad2c2-9d3b-4990-92c7-117cf25fdae3",
		Model:  "doubao-seed-1-6-251015"}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

// newChatModel1 component initialization function of node 'ChatModel6' in graph 'mytest2'
func newChatModel1(ctx context.Context) (cm model.ChatModel, err error) {
	// TODO Modify component configuration here.
	config := &ark.ChatModelConfig{
		APIKey: "358ad2c2-9d3b-4990-92c7-117cf25fdae3",
		Model:  "doubao-seed-1-6-251015"}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
