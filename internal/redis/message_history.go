package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/herman-xphp/go-chat-server/pkg/models"
)

const (
	messageHistoryKey = "chat:messages"
	maxHistorySize    = 100
)

type MessageHistory struct {
	client *Client
}

func NewMessageHistory(client *Client) *MessageHistory {
	return &MessageHistory{client: client}
}

func (m *MessageHistory) AddMessage(ctx context.Context, message models.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	pipe := m.client.Pipeline()
	pipe.LPush(ctx, messageHistoryKey, data)
	pipe.LTrim(ctx, messageHistoryKey, 0, maxHistorySize-1)
	pipe.Expire(ctx, messageHistoryKey, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

func (m *MessageHistory) GetRecentMessages(ctx context.Context, limit int64) ([]models.Message, error) {
	if limit <= 0 || limit > maxHistorySize {
		limit = maxHistorySize
	}

	result, err := m.client.LRange(ctx, messageHistoryKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0, len(result))
	for i := len(result) - 1; i >= 0; i-- {
		var msg models.Message
		if err := json.Unmarshal([]byte(result[i]), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
