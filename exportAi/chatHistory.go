package exportAi

import (
	"context"
	"encoding/json"
	"fmt"
	"main/database"

	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

type ChatMessage struct {
	MsgID     string          `json:"msg_id"`
	Role      schema.RoleType `json:"role"`
	Content   string          `json:"content"`
	Model     string          `json:"model"`
	Timestamp int64           `json:"timestamp"`
}

func (c *ChatMessage) SaveChatMessage(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("chat:%s", sessionID)
	msgJSON, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return database.RedisDb.ZAdd(ctx, key, redis.Z{
		Score:  float64(c.Timestamp),
		Member: string(msgJSON),
	}).Err()
}

func (c *ChatMessage) GetChatHistory(ctx context.Context, sessionID string, pageSize int, offset int64) ([]ChatMessage, error) {
	key := fmt.Sprintf("chat:%s", sessionID)
	res, err := database.RedisDb.ZRange(ctx, key, offset, offset+int64(pageSize)-1).Result()
	if err != nil {
		return nil, err
	}
	msgs := make([]ChatMessage, 0, len(res))
	for _, msgJSON := range res {
		var msg ChatMessage
		if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
			continue
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}
