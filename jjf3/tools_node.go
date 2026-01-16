package jjf3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	return &schema.ToolInfo{
		Name: "calculate",
		Desc: "这是一个加减工具，计算两个数的结果",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"number1": {
				Type:     schema.Integer,
				Required: true,
			},
			"number2": {
				Type:     schema.Integer,
				Required: true,
			},
		}),
	}, nil

}

func (impl *ToolImpl) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if argumentsInJSON == "" {
		return "", errors.New("argument is empty")
	}
	arg := struct {
		Number  int `json:"number1"`
		Number2 int `json:"number2"`
	}{}
	_ = json.Unmarshal([]byte(argumentsInJSON), &arg)
	fmts := fmt.Sprintf("结果是：%d", arg.Number+arg.Number2)
	return fmts, nil
}
