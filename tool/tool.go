package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	react2 "github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/xuri/excelize/v2"
)

type UserPlayer struct {
	ID        int    `json:"id"`
	Uid       int    `json:"uid"`
	GiftId    int    `json:"gift_id"`
	GiftName  string `json:"gift_name"`
	GiftImage string `json:"gift_image"`
}

type GetDbData interface {
	GetAll() (string, error)
}
type ExportDBDataTool struct{}

func (p UserPlayer) GetAll() (string, error) {

	// 直接硬编码10条数据，每条数据值固定且唯一
	ss := []*UserPlayer{
		// 第1条数据
		{
			ID:        1,
			Uid:       10001,
			GiftId:    1,
			GiftName:  "金币礼包",
			GiftImage: "https://example.com/gift1.png",
		},
		// 第2条数据
		{
			ID:        2,
			Uid:       10002,
			GiftId:    2,
			GiftName:  "钻石礼包",
			GiftImage: "https://example.com/gift2.png",
		},
		// 第3条数据
		{
			ID:        3,
			Uid:       10003,
			GiftId:    3,
			GiftName:  "经验礼包",
			GiftImage: "https://example.com/gift3.png",
		},
		// 第4条数据
		{
			ID:        4,
			Uid:       10004,
			GiftId:    4,
			GiftName:  "道具礼包",
			GiftImage: "https://example.com/gift4.png",
		},
		// 第5条数据
		{
			ID:        5,
			Uid:       10005,
			GiftId:    5,
			GiftName:  "稀有礼包",
			GiftImage: "https://example.com/gift5.png",
		},
		// 第6条数据
		{
			ID:        6,
			Uid:       10006,
			GiftId:    1,
			GiftName:  "金币礼包",
			GiftImage: "https://example.com/gift1.png",
		},
		// 第7条数据
		{
			ID:        7,
			Uid:       10007,
			GiftId:    2,
			GiftName:  "钻石礼包",
			GiftImage: "https://example.com/gift2.png",
		},
		// 第8条数据
		{
			ID:        8,
			Uid:       10008,
			GiftId:    3,
			GiftName:  "经验礼包",
			GiftImage: "https://example.com/gift3.png",
		},
		// 第9条数据
		{
			ID:        9,
			Uid:       10009,
			GiftId:    4,
			GiftName:  "道具礼包",
			GiftImage: "https://example.com/gift4.png",
		},
		// 第10条数据
		{
			ID:        10,
			Uid:       10010,
			GiftId:    5,
			GiftName:  "稀有礼包",
			GiftImage: "https://example.com/gift5.png",
		},
	}
	js, err := json.Marshal(ss)
	if err != nil {
		return "", err
	}
	return string(js), nil
}

var tableFieldMeta = map[string]map[string]string{
	"user_player": {
		"id":         "自增id",
		"uid":        "用户的id",
		"gift_id":    "礼物的id",
		"gift_name":  "礼物的名字",
		"gift_image": "礼物的图片",
	},
}

func getValidFields(tableName string) []string {
	fields := make([]string, 0)
	if meta, ok := tableFieldMeta[tableName]; ok {
		for field := range meta {
			fields = append(fields, field)
		}
	}
	return fields
}
func genFieldPrompt() string {
	prompt := "以下是所有表的合法字段（必须严格使用这些字段名）：\n"
	for table, fields := range tableFieldMeta {
		prompt += fmt.Sprintf("- %s表：%s（字段说明：%v）\n", table, strings.Join(getValidFields(table), ","), fields)
	}
	prompt += `规则：
1. 用户说“导出用户表的姓名和年龄”，则table_name=user，fields=name,age；
2. 用户说“导出订单表的订单ID和金额”，则table_name=order，fields=order_id,amount；
3. 字段名必须严格匹配，不能自创（如user表没有username，只有name）；
4. 若用户说的字段不存在，提示用户该表的合法字段；
5. 如果有自增id则默认导出都包含改字段。`
	return prompt
}

var modelMap = map[string]GetDbData{
	"user_player": UserPlayer{}, // "user"表对应UserModel
}

func (u *ExportDBDataTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_db_data",
		Desc: "导出数据库数据",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"table_name": {
				Type:     schema.String,
				Required: true,
				Desc:     "数据库表名",
				Enum:     []string{"user_player"},
			},
			"fields": {
				Type:     schema.String,
				Required: false,
				Desc:     "要导出的字段名列表，多个字段用英文逗号分隔（如name,age），为空则导出所有字段",
			},
		}),
	}, nil
}

func (u *ExportDBDataTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var args struct {
		TableName string `json:"table_name"`
		Fields    int    `json:"fields"`
	}
	json.Unmarshal([]byte(argumentsInJSON), &args)
	model := modelMap[args.TableName]
	data, err := model.GetAll()
	if err != nil {
		return "", err
	}
	userResult, _ := json.Marshal(data)
	return string(userResult), nil
}

func saveExcel(headers []string, data [][]interface{}) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("关闭 Excel 文件失败: %v", err)
		}
	}()
	sheetName := "用户信息表"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatalf("创建工作表失败: %v", err)
	}
	f.SetActiveSheet(index)
	for col, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+col, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			log.Fatalf("写入表头失败（列%c）：%v", 'A'+col, err)
		}
	}
	for rowIdx, rowData := range data {
		rowNum := rowIdx + 2
		for colIdx, value := range rowData {
			cell := fmt.Sprintf("%c%d", 'A'+colIdx, rowNum)
			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				log.Fatalf("写入数据失败（行%d列%c）：%v", rowNum, 'A'+colIdx, err)
			}
		}
	}

	filePath := "./用户信息.xlsx"
	if err := f.SaveAs(filePath); err != nil {
		log.Fatalf("保存 Excel 文件失败: %v", err)
	}

	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("Excel 文件已成功保存到: %s\n", filePath)
	} else {
		log.Fatalf("验证文件失败: %v", err)
	}
}

type Import struct{}

func (u *Import) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "export",
		Desc: "导出数据",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"headers": {
				Type:     schema.String,
				Required: true,
			},
			"data": {
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}
func (u *Import) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var args struct {
		HeadersStr string `json:"headers"`
		DataStr    string `json:"data"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析工具参数失败：%v", err)
	}

	headers := strings.Split(args.HeadersStr, ",")

	var mapSlice []map[string]interface{}
	err := json.Unmarshal([]byte(args.DataStr), &mapSlice)
	if err != nil {
		log.Fatal("解析 JSON 失败：", err)
	}
	var data [][]interface{}
	for _, item := range mapSlice {
		var row []interface{}
		for _, field := range headers {
			row = append(row, item[field]) // 按字段顺序提取值
		}
		data = append(data, row)
	}

	saveExcel(headers, data)

	return "保存成功", nil
}

func Tool(ctx context.Context) {

	toolList := []tool.BaseTool{
		&ExportDBDataTool{},
		&Import{},
	}
	//toolMod, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: toolList})
	//if err != nil {
	//	panic(err)
	//}
	//calculInfo, err := toolList[1].Info(ctx) // toolList[1] 是 &Calculator{}
	//if err != nil {
	//	panic(err)
	//}
	//userInfo, err := toolList[0].Info(ctx)
	//if err != nil {
	//	panic(err)
	//}

	//toolInfo := []*schema.ToolInfo{calculInfo, userInfo}

	m, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: "358ad2c2-9d3b-4990-92c7-117cf25fdae3",
		Model:  "doubao-seed-1-6-251015",
	})
	if err != nil {
		panic(err)
		return
	}
	message := []*schema.Message{
		schema.SystemMessage(genFieldPrompt()),
		schema.UserMessage("帮我导出user_player表数据 但我只要礼物名字数据"),
	}

	react, err := react2.NewAgent(ctx, &react2.AgentConfig{
		ToolCallingModel: m,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: toolList},
	})
	if err != nil {
		panic(err)
	}

	response, err := react.Generate(ctx, message)
	if err != nil {
		panic(err)
	}
	fmt.Println(response.Content)

	//response, err := m.Generate(ctx, message, model.WithTools(toolInfo))
	//if err != nil {
	//	panic(err)
	//}
	if len(response.ToolCalls) > 0 {
		for _, call := range response.ToolCalls {
			fmt.Printf("使用工具%s", call.Function.Name)
			fmt.Printf("参数%s", call.Function.Arguments)
			//toolRes, err := toolMod.Invoke(ctx, response)
			//if err != nil {
			//	panic(err)
			//}
			fmt.Printf("工具返回结果%s", call.Extra)
		}
	}
}
