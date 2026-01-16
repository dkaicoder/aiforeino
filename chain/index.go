package chain

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func Mains() {
	ctx := context.Background()
	g := compose.NewGraph[map[string]string, *schema.Message]()
	Lambda := compose.InvokableLambda(func(ctx context.Context, input map[string]string) (output map[string]string, err error) {
		if input["rule"] == "cute" {
			output = map[string]string{"rule": "cute", "content": input["content"]}
		}
		if input["rule"] == "angry" {
			output = map[string]string{"rule": "angry", "content": input["content"]}
		}
		return
	})

	cuteLambda := compose.InvokableLambda(func(ctx context.Context, input map[string]string) (output []*schema.Message, err error) {
		return []*schema.Message{
			{
				Role:    schema.System,
				Content: "你是一个可爱的小女孩，每次都会用可爱的语气回答我的问题",
			},
			{
				Role:    schema.User,
				Content: input["content"],
			},
		}, nil
	})

	angryLambda := compose.InvokableLambda(func(ctx context.Context, input map[string]string) (output []*schema.Message, err error) {
		return []*schema.Message{
			{
				Role:    schema.System,
				Content: "你是一个生气的小女孩，每次都会用生气的语气回答我的问题",
			},
			{
				Role:    schema.User,
				Content: input["content"],
			},
		}, nil
	})

	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: "358ad2c2-9d3b-4990-92c7-117cf25fdae3",
		Model:  "doubao-seed-1-6-251015",
	})
	if err != nil {
		panic(err)
	}
	_ = g.AddLambdaNode("lambda", Lambda)
	_ = g.AddLambdaNode("cute", cuteLambda)
	_ = g.AddLambdaNode("angry", angryLambda)
	_ = g.AddChatModelNode("model", model)

	_ = g.AddEdge(compose.START, "lambda")
	_ = g.AddEdge("cute", "model")
	_ = g.AddEdge("angry", "model")
	_ = g.AddEdge("model", compose.END)

	_ = g.AddBranch("lambda", compose.NewGraphBranch(func(ctx context.Context, in map[string]string) (string, error) {
		if in["rule"] == "cute" {
			return "cute", nil
		}
		if in["rule"] == "angry" {
			return "angry", nil
		}
		return "cute", nil
	}, map[string]bool{"cute": true, "angry": true}))

	//编译
	r, err := g.Compile(ctx)
	if err != nil {
		panic(err)
	}
	input := map[string]string{"rule": "angry", "content": "你好，我是小明，你喜欢吗？"}
	//执行
	answer, err := r.Invoke(ctx, input)
	if err != nil {
		panic(err)
	}
	fmt.Println(answer)

}
