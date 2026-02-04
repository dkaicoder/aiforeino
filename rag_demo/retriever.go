package rag_demo

import (
	"context"

	redisRet "github.com/cloudwego/eino-ext/components/retriever/redis"
)

func (r *RAGEngine) newRetriever(ctx context.Context) {
	re, err := redisRet.NewRetriever(ctx, &redisRet.RetrieverConfig{
		Client:            r.redis,
		Index:             r.indexName,
		VectorField:       "vector_content",
		DistanceThreshold: nil,
		Dialect:           2,
		ReturnFields:      []string{"vector_content", "content"},
		DocumentConverter: nil,
		TopK:              5,
		Embedding:         r.embedder,
	})

	if err != nil {
		r.Err = err
		return
	}

	r.Retriever = re
}
