package redis

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/herman-xphp/go-chat-server/pkg/models"
)

const (
	onlineUsersKey = "chat:online_users"
	userTTL        = 30 * time.Second
)

type Presence struct {
	client *Client
}

func NewPresence(client *Client) *Presence {
	return &Presence{client: client}
}

func (p *Presence) SetUserOnline(ctx context.Context, user models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return p.client.Set(ctx, userKey(user.ID), data, userTTL).Err()
}

func (p *Presence) SetUserOffline(ctx context.Context, userID string) error {
	return p.client.Del(ctx, userKey(userID)).Err()
}

func (p *Presence) GetOnlineUsers(ctx context.Context) ([]models.User, error) {
	keys, err := p.client.Keys(ctx, "chat:user:*").Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []models.User{}, nil
	}

	pipe := p.client.Pipeline()
	cmds := make([]*goredis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != goredis.Nil {
		return nil, err
	}

	users := make([]models.User, 0, len(cmds))
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			continue
		}

		var user models.User
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (p *Presence) KeepAlive(ctx context.Context, userID string) error {
	key := userKey(userID)
	return p.client.Expire(ctx, key, userTTL).Err()
}

func userKey(userID string) string {
	return "chat:user:" + userID
}
