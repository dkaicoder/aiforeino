package llm

import (
	"context"
	"main/config"

	"github.com/cloudwego/eino-ext/components/model/ark"
)

func NewChatModelFactory(ctx context.Context, modelName string) (*ark.ChatModel, error) {
	cfg := config.GetConfig()
	chatConfig := &ark.ChatModelConfig{
		APIKey: cfg.ApiKey,
		Model:  modelName,
	}
	// 公共实例化逻辑
	chatModel, err := ark.NewChatModel(ctx, chatConfig)
	if err != nil {
		return nil, err
	}
	return chatModel, nil
}
