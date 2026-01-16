package jjf

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// newToolsNode component initialization function of node 'ToolsNode1' in graph 'mytest'
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
	return &schema.ToolInfo{}, nil
}

func (impl *ToolImpl) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	//fmt.Println("123123123123")
	//var args struct {
	//	Number1 int `json:"numbe1"`
	//	Number2 int `json:"numbe2"`
	//}
	//// 补全错误处理（之前缺失）
	//if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
	//	return "", err
	//}
	//
	//num := args.Number1 + args.Number2
	//// 返回纯数字字符串（不是 JSON）
	//fmt.Println("123123123123", num)
	return argumentsInJSON, nil
}
