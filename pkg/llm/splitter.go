package llm

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/components/document"
)

func NewSplitter(ctx context.Context) (document.Transformer, error) {
	t, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"#": "title",
		},
		TrimHeaders: false,
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}
