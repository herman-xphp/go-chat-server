package hub

import (
	"log"
	"time"

	"github.com/herman-xphp/go-chat-server/pkg/models"
)

func (h *Hub) sendMessageHistory(client Client) {
	messages, err := h.messageHistory.GetRecentMessages(h.ctx, 50)
	if err != nil {
		log.Printf("Failed to get message history: %v", err)
		return
	}

	for _, msg := range messages {
		client.Send(msg)
	}
}

func (h *Hub) sendUserList(client Client) {
	users, err := h.presence.GetOnlineUsers(h.ctx)
	if err != nil {
		log.Printf("Failed to get online users: %v", err)
		return
	}

	usernames := make([]string, 0, len(users))
	for _, user := range users {
		usernames = append(usernames, user.Username)
	}

	client.Send(models.Message{
		Type:    models.MessageTypeUserList,
		Content: "",
	})
}

func (h *Hub) startPresenceHeartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.updatePresence()
		}
	}
}

func (h *Hub) updatePresence() {
	h.mutex.RLock()
	clients := make([]Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mutex.RUnlock()

	for _, client := range clients {
		if err := h.presence.KeepAlive(h.ctx, client.GetID()); err != nil {
			log.Printf("Failed to update presence for %s: %v", client.GetUsername(), err)
		}
	}
}

func (h *Hub) Shutdown() {
	h.cancel()
	if h.pubsub != nil {
		h.pubsub.Close()
	}
}
