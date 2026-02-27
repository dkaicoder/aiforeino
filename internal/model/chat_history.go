package model

import (
	"github.com/cloudwego/eino/schema"
)

type ChatMessage struct {
	MsgID     string          `json:"msg_id"`
	Role      schema.RoleType `json:"role"`
	Content   string          `json:"content"`
	Model     string          `json:"model"`
	Timestamp int64           `json:"timestamp"`
}
