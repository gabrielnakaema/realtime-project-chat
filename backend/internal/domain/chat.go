package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChatType string

var (
	ChatTypeGeneral ChatType = "general"
	ChatTypeProject ChatType = "project"
)

const ChatUnreadCountFetchLimit = 100

type Chat struct {
	Id            uuid.UUID  `json:"id"`
	ProjectId     *uuid.UUID `json:"project_id"`
	ChatType      ChatType   `json:"chat_type"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	UnreadCount   int        `json:"unread_count"`
	HasMoreUnread bool       `json:"has_more_unread"`

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
	ReadsCount  int         `json:"reads_count"`

	Member *ChatMember `json:"member,omitempty"`
}

type ChatMessageRead struct {
	MessageId uuid.UUID `json:"message_id"`
	UserId    uuid.UUID `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`

	User *User `json:"user,omitempty"`
}

type ChatUnreadSummary struct {
	UnreadCount   int  `json:"unread_count"`
	HasMoreUnread bool `json:"has_more_unread"`
}
