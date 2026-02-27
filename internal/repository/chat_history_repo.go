package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/database"
	"main/internal/model"

	"github.com/redis/go-redis/v9"
)

type ChatHistoryRepository interface {
	SaveChatMessage(ctx context.Context, sessionID string, msg *model.ChatMessage) error
	GetChatHistory(ctx context.Context, sessionID string, pageSize int, offset int64) ([]*model.ChatMessage, error)
}

type redisChatHistoryRepo struct {
	client *redis.Client
}

func NewRedisChatHistoryRepo(client *redis.Client) ChatHistoryRepository {
	return &redisChatHistoryRepo{client: client}
}

func (c *redisChatHistoryRepo) SaveChatMessage(ctx context.Context, sessionID string, msg *model.ChatMessage) error {
	key := fmt.Sprintf("chat:%s", sessionID)
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(msg.Timestamp),
		Member: string(msgJSON),
	}).Err()
}

func (c *redisChatHistoryRepo) GetChatHistory(ctx context.Context, sessionID string, pageSize int, offset int64) ([]*model.ChatMessage, error) {
	key := fmt.Sprintf("chat:%s", sessionID)
	res, err := database.RedisDb.ZRevRange(ctx, key, offset, offset+int64(pageSize)-1).Result()
	if err != nil {
		return nil, err
	}

	msgs := make([]*model.ChatMessage, 0, len(res))
	for _, msgJSON := range res {
		var msg *model.ChatMessage
		if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
			continue
		}
		msgs = append(msgs, msg)
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
