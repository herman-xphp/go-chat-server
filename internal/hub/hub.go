package hub

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/herman-xphp/go-chat-server/internal/redis"
	"github.com/herman-xphp/go-chat-server/pkg/models"
)

type Client interface {
	GetID() string
	GetUsername() string
	Send(message models.Message)
	Close()
}

type Hub struct {
	clients        map[string]Client
	broadcast      chan models.Message
	register       chan Client
	unregister     chan Client
	mutex          sync.RWMutex
	redisClient    *redis.Client
	messageHistory *redis.MessageHistory
	presence       *redis.Presence
	rateLimiter    *redis.RateLimiter
	pubsub         *redis.PubSub
	ctx            context.Context
	cancel         context.CancelFunc
	instanceID     string // Unique ID for this server instance
}

func NewHub(redisClient *redis.Client, instanceID string) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	h := &Hub{
		clients:        make(map[string]Client),
		broadcast:      make(chan models.Message, 256),
		register:       make(chan Client),
		unregister:     make(chan Client),
		redisClient:    redisClient,
		messageHistory: redis.NewMessageHistory(redisClient),
		presence:       redis.NewPresence(redisClient),
		rateLimiter:    redis.NewRateLimiter(redisClient),
		pubsub:         redis.NewPubSub(redisClient),
		ctx:            ctx,
		cancel:         cancel,
		instanceID:     instanceID,
	}

	return h
}

func (h *Hub) Run() {
	// Subscribe to Redis Pub/Sub for multi-instance support
	redisChan := h.pubsub.Subscribe(h.ctx)

	// Start presence heartbeat
	go h.startPresenceHeartbeat()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)

		case message := <-redisChan:
			// Message from other server instances via Redis Pub/Sub
			// Skip if it's from this instance (already broadcasted locally)
			if message.InstanceID != h.instanceID {
				h.broadcastToLocalClients(message)
			}
		}
	}
}

func (h *Hub) registerClient(client Client) {
	h.mutex.Lock()
	h.clients[client.GetID()] = client
	h.mutex.Unlock()

	log.Printf("Client registered: %s (%s)", client.GetUsername(), client.GetID())

	// Register user in Redis presence
	user := models.User{
		ID:       client.GetID(),
		Username: client.GetUsername(),
		IsOnline: true,
		JoinedAt: time.Now(),
	}
	if err := h.presence.SetUserOnline(h.ctx, user); err != nil {
		log.Printf("Failed to set user online in Redis: %v", err)
	}

	// Send message history to new client
	h.sendMessageHistory(client)

	// Broadcast user joined
	joinMsg := models.Message{
		Type:      models.MessageTypeJoin,
		From:      client.GetUsername(),
		Content:   client.GetUsername() + " joined the chat",
		Timestamp: time.Now(),
	}
	h.broadcast <- joinMsg

	// Send user list to the new client
	h.sendUserList(client)
}

func (h *Hub) unregisterClient(client Client) {
	h.mutex.Lock()
	if _, ok := h.clients[client.GetID()]; ok {
		delete(h.clients, client.GetID())
		client.Close()
	}
	h.mutex.Unlock()

	log.Printf("Client unregistered: %s (%s)", client.GetUsername(), client.GetID())

	// Remove from Redis presence
	if err := h.presence.SetUserOffline(h.ctx, client.GetID()); err != nil {
		log.Printf("Failed to set user offline in Redis: %v", err)
	}

	// Broadcast user left
	leaveMsg := models.Message{
		Type:      models.MessageTypeLeave,
		From:      client.GetUsername(),
		Content:   client.GetUsername() + " left the chat",
		Timestamp: time.Now(),
	}
	h.broadcast <- leaveMsg
}

func (h *Hub) handleBroadcast(message models.Message) {
	// Check rate limit for broadcast messages
	if message.Type == models.MessageTypeBroadcast {
		// Find client by username to get ID
		var clientID string
		h.mutex.RLock()
		for _, client := range h.clients {
			if client.GetUsername() == message.From {
				clientID = client.GetID()
				break
			}
		}
		h.mutex.RUnlock()

		if clientID != "" {
			allowed, err := h.rateLimiter.AllowMessage(h.ctx, clientID)
			if err != nil {
				log.Printf("Rate limit check error: %v", err)
			} else if !allowed {
				log.Printf("Rate limit exceeded for user: %s", message.From)
				return
			}
		}
	}

	// Store message in history
	if message.Type == models.MessageTypeBroadcast || message.Type == models.MessageTypePrivate {
		if err := h.messageHistory.AddMessage(h.ctx, message); err != nil {
			log.Printf("Failed to store message in history: %v", err)
		}
	}

	// Add instance ID to message
	message.InstanceID = h.instanceID

	// Broadcast locally FIRST for instant response
	h.broadcastToLocalClients(message)

	// Then publish to Redis Pub/Sub for other instances (non-blocking)
	go func() {
		if err := h.pubsub.Publish(h.ctx, message); err != nil {
			log.Printf("Failed to publish to Redis: %v", err)
		}
	}()
}

func (h *Hub) broadcastToLocalClients(message models.Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if message.Type == models.MessageTypePrivate {
		// Private message: send only to recipient
		if recipient, ok := h.clients[message.To]; ok {
			recipient.Send(message)
		}
		// Also send to sender for confirmation
		if sender, ok := h.clients[message.From]; ok {
			sender.Send(message)
		}
	} else {
		// Broadcast to all clients
		for _, client := range h.clients {
			client.Send(message)
		}
	}
}

func (h *Hub) Register(client Client) {
	h.register <- client
}

func (h *Hub) Unregister(client Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(message models.Message) {
	h.broadcast <- message
}
