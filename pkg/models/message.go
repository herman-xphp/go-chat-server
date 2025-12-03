package models

import "time"

type MessageType string

const (
	MessageTypeBroadcast MessageType = "broadcast"
	MessageTypePrivate   MessageType = "private"
	MessageTypeJoin      MessageType = "join"
	MessageTypeLeave     MessageType = "leave"
	MessageTypeUserList  MessageType = "user_list"
)

type Message struct {
	Type       MessageType `json:"type"`
	From       string      `json:"from"`
	To         string      `json:"to,omitempty"`
	Content    string      `json:"content"`
	Timestamp  time.Time   `json:"timestamp"`
	Room       string      `json:"room,omitempty"`
	InstanceID string      `json:"instance_id,omitempty"` // For multi-instance dedup
}

type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	IsOnline bool      `json:"is_online"`
	JoinedAt time.Time `json:"joined_at"`
}
