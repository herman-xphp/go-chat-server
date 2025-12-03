package redis

import (
	"context"
	"fmt"
	"time"
)

const (
	rateLimitWindow   = 10 * time.Second
	maxMessagesPerWindow = 10
)

type RateLimiter struct {
	client *Client
}

func NewRateLimiter(client *Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (r *RateLimiter) AllowMessage(ctx context.Context, userID string) (bool, error) {
	key := rateLimitKey(userID)

	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		r.client.Expire(ctx, key, rateLimitWindow)
	}

	return count <= maxMessagesPerWindow, nil
}

func (r *RateLimiter) GetRemainingMessages(ctx context.Context, userID string) (int64, error) {
	key := rateLimitKey(userID)

	count, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if err.Error() == "redis: nil" {
			return maxMessagesPerWindow, nil
		}
		return 0, err
	}

	remaining := maxMessagesPerWindow - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

func (r *RateLimiter) ResetUserLimit(ctx context.Context, userID string) error {
	key := rateLimitKey(userID)
	return r.client.Del(ctx, key).Err()
}

func rateLimitKey(userID string) string {
	return fmt.Sprintf("chat:ratelimit:%s", userID)
}
