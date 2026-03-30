package llm

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
)

func NewLoader(ctx context.Context) (*file.FileLoader, error) {
	l, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
		Parser:      nil,
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}
