package exportAi

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// newBranch branch initialization method of node 'ChatModel1' in graph 'agent'
func newBranch(ctx context.Context, input []*schema.Message) (endNode string, err error) {
	if !IsStartWithBraceByRegex(input[0].Content) {
		return compose.END, nil
	}
	graphChoice := GraphChoice{}
	json.Unmarshal([]byte(input[0].Content), &graphChoice)
	if graphChoice.GraphType == "" {
		return "", errors.New("graph_type is empty")
	}
	if graphChoice.GraphType == "export" {
		return "Graph1", nil
	} else {
		return compose.END, nil
	}

}
func IsStartWithBraceByRegex(s string) bool {
	pattern := regexp.MustCompile(`^\{`)
	return pattern.MatchString(s)
}
