package main

import (
	"context"
	"github.com/cloudwego/eino-ext/components/document/loader/file"
)

func (r *RAGEngine) newLoader(ctx context.Context) {
	l, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
		Parser:      nil,
	})
	if err != nil {
		r.Err = err
		return
	}
	r.Loader = l
}
