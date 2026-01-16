package jjf5

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// newToolsNode component initialization function of node 'ToolsNode1' in graph 'mytest2'
func newToolsNode(ctx context.Context) (tsn *compose.ToolsNode, err error) {
	// TODO Modify component configuration here.
	config := &compose.ToolsNodeConfig{}
	toolIns11, err := newTool(ctx)
	if err != nil {
		return nil, err
	}
	config.Tools = []tool.BaseTool{toolIns11}
	tsn, err = compose.NewToolNode(ctx, config)
	if err != nil {
		return nil, err
	}
	return tsn, nil
}

type ToolImpl struct {
	config *ToolConfig
}

type ToolConfig struct {
}

func newTool(ctx context.Context) (bt tool.BaseTool, err error) {
	// TODO Modify component configuration here.
	config := &ToolConfig{}
	bt = &ToolImpl{config: config}
	return bt, nil
}

func (impl *ToolImpl) Info(ctx context.Context) (*schema.ToolInfo, error) {
	panic("implement me")
}

func (impl *ToolImpl) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	panic("implement me")
}
