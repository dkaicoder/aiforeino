package export_graph

import (
	"context"
	"fmt"
	"main/internal/database"

	"github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/retriever"
)

// newRetriever component initialization function of node 'Retriever1' in graph 'mytest2'
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	initRedis := database.RedisDb
	embeddingIns11, err := newEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	config := &redis.RetrieverConfig{
		Client:            initRedis,
		Index:             "OuterIndex",
		VectorField:       "vector_content",
		Dialect:           2,
		ReturnFields:      []string{"content"}, // 确保 Redis Hash 里有这两个字段
		TopK:              5,
		DocumentConverter: nil,
		DistanceThreshold: nil,
		Embedding:         embeddingIns11,
	}
	rtr, err = redis.NewRetriever(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("初始化Retriever失败: %w", err)
	}
	return rtr, nil
}
