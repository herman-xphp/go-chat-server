package redis

import (
	"context"
	"encoding/json"
	"log"

	goredis "github.com/redis/go-redis/v9"
	"github.com/herman-xphp/go-chat-server/pkg/models"
)

const (
	chatChannelName = "chat:messages"
)

type PubSub struct {
	client *Client
	pubsub *goredis.PubSub
}

func NewPubSub(client *Client) *PubSub {
	return &PubSub{
		client: client,
	}
}

func (p *PubSub) Subscribe(ctx context.Context) <-chan models.Message {
	p.pubsub = p.client.Subscribe(ctx, chatChannelName)

	messageChan := make(chan models.Message, 100)

	go func() {
		defer close(messageChan)
		defer p.pubsub.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-p.pubsub.Channel():
				var message models.Message
				if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					continue
				}
				messageChan <- message
			}
		}
	}()

	return messageChan
}

func (p *PubSub) Publish(ctx context.Context, message models.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, chatChannelName, data).Err()
}

func (p *PubSub) Close() error {
	if p.pubsub != nil {
		return p.pubsub.Close()
	}
	return nil
}
