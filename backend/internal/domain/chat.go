package domain

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	Id        uuid.UUID `json:"id"`
	ProjectId uuid.UUID `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Members  []ChatMember  `json:"members,omitempty"`
	Messages []ChatMessage `json:"messages,omitempty"`
}

type ChatMember struct {
	ChatId     uuid.UUID `json:"chat_id,omitempty"`
	UserId     uuid.UUID `json:"user_id,omitempty"`
	LastSeenAt time.Time `json:"last_seen_at,omitempty"`
	JoinedAt   time.Time `json:"joined_at,omitempty"`

	User *User `json:"user,omitempty"`
	Chat *Chat `json:"chat,omitempty"`
}

type MessageType string

var (
	MessageTypeText   MessageType = "text"
	MessageTypeSystem MessageType = "system"
)

type ChatMessage struct {
	Id          uuid.UUID   `json:"id"`
	ChatId      uuid.UUID   `json:"chat_id"`
	UserId      *uuid.UUID  `json:"user_id"` // If UserId is nil, the message is a system message
	MessageType MessageType `json:"message_type"`
	Content     string      `json:"content"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`

	Member *ChatMember `json:"member,omitempty"`
}
