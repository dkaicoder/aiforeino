package main

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/ark"
)

func (r *RAGEngine) newChatModel(ctx context.Context) {
	m, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: "358ad2c2-9d3b-4990-92c7-117cf25fdae3",
		Model:  "doubao-seed-1-6-251015",
	})
	if err != nil {
		r.Err = err
		return
	}

	r.ChatModel = m
}
