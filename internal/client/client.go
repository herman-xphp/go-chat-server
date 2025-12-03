package client

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/herman-xphp/go-chat-server/internal/hub"
	"github.com/herman-xphp/go-chat-server/pkg/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	id       string
	username string
	hub      *hub.Hub
	conn     *websocket.Conn
	send     chan models.Message
}

func NewClient(id, username string, hub *hub.Hub, conn *websocket.Conn) *Client {
	return &Client{
		id:       id,
		username: username,
		hub:      hub,
		conn:     conn,
		send:     make(chan models.Message, 256),
	}
}

func (c *Client) GetID() string {
	return c.id
}

func (c *Client) GetUsername() string {
	return c.username
}

func (c *Client) Send(message models.Message) {
	select {
	case c.send <- message:
	default:
		close(c.send)
	}
}

func (c *Client) Close() {
	close(c.send)
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg models.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		msg.From = c.username
		msg.Timestamp = time.Now()

		c.hub.Broadcast(msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			encoder := json.NewEncoder(w)
			if err := encoder.Encode(message); err != nil {
				return
			}

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				msg := <-c.send
				encoder.Encode(msg)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
