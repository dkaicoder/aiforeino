package llm

import (
	"context"
	"main/internal/database"

	redisRet "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/embedding"
)

func NewRetriever(ctx context.Context, indexName string, embedder embedding.Embedder) (retriever *redisRet.Retriever, err error) {
	re, err := redisRet.NewRetriever(ctx, &redisRet.RetrieverConfig{
		Client:            database.RedisDb,
		Index:             indexName,
		VectorField:       "vector_content",
		DistanceThreshold: nil,
		Dialect:           2,
		ReturnFields:      []string{"vector_content", "content"},
		DocumentConverter: nil,
		TopK:              5,
		Embedding:         embedder,
	})
	if err != nil {
		return nil, err
	}

	return re, nil
}
