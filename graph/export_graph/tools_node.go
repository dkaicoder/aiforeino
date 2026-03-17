package export_graph

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/database"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// newToolsNode component initialization function of node 'ToolsNode1' in graph 'mytest2'
func newToolsNode(ctx context.Context) (tsn *compose.ToolsNode, err error) {
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
		Name: "sql_verifier",
		Desc: "这是一个SQL校验工具，用于验证生成的SQL是否语法合法、逻辑正确、符合表结构",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"sql": {
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

func (impl *ToolImpl) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	sqlStruct := struct {
		SQL string `json:"sql"`
	}{}
	res := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   string `json:"data"`
	}{}
	json.Unmarshal([]byte(argumentsInJSON), &sqlStruct)
	db := database.InitMysql(ctx)
	sql := fmt.Sprintf("EXPLAIN %s", sqlStruct.SQL)
	fmt.Println(sql)
	err := db.Raw(sql).Error
	if err != nil {
		return "", fmt.Errorf("SQL 语法错误: %v\n", err)
	}
	res.Status = 200
	res.Msg = "success"
	res.Data = sqlStruct.SQL
	js, _ := json.Marshal(res)
	return string(js), nil
}
