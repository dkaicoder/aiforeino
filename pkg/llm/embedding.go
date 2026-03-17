package llm

import (
	"context"
	"main/config"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/embedding"
)

func NewEmbeddingFactory(ctx context.Context, model string) (eb embedding.Embedder, err error) {
	cfg := config.GetConfig()
	embeddingConfig := &ark.EmbeddingConfig{
		APIKey: cfg.Embedding.ApiKey,
		Model:  model}
	eb, err = ark.NewEmbedder(ctx, embeddingConfig)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
