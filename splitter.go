package main

import (
	"context"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
)

func (r *RAGEngine) newSplitter(ctx context.Context) {
	t, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"#": "title",
		},
		TrimHeaders: false,
	})
	if err != nil {
		r.Err = err
		return
	}
	r.Splitter = t
}
