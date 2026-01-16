package jjf5

import (
	"context"

	"github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/retriever"
)

// newRetriever component initialization function of node 'Retriever1' in graph 'mytest2'
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	// TODO Modify component configuration here.
	config := &redis.RetrieverConfig{
		Index:        "OuterIndex",
		VectorField:  "vector_content",
		Dialect:      2,
		ReturnFields: []string{"vector_content", "content"},
		TopK:         2}
	embeddingIns11, err := newEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	config.Embedding = embeddingIns11
	rtr, err = redis.NewRetriever(ctx, config)
	if err != nil {
		return nil, err
	}
	return rtr, nil
}
