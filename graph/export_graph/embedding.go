package export_graph

import (
	"context"
	"main/pkg/llm"

	"github.com/cloudwego/eino/components/embedding"
)

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	em, err := llm.NewEmbeddingFactory(ctx, "doubao-embedding-text-240715")
	if err != nil {
		return nil, err
	}
	return em, nil
}
