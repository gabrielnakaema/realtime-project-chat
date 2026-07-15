package events

import (
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
)

type NotificationCreatedPayload struct {
	Notification domain.Notification `json:"notification"`
}

type TaskCreatedPayload struct {
	Task         domain.Task         `json:"task"`
	User         domain.User         `json:"user"`
	ActionOrigin domain.ActionOrigin `json:"action_origin,omitempty"`
}

type TaskUpdatedPayload struct {
	Task                    domain.Task         `json:"task"`
	PreviousTask            *domain.Task        `json:"previous_task,omitempty"`
	User                    domain.User         `json:"user"`
	PreviousProjectColumnID *uuid.UUID          `json:"previous_project_column_id,omitempty"`
	ActionOrigin            domain.ActionOrigin `json:"action_origin,omitempty"`
}

type TaskCommentCreatedPayload struct {
	TaskComment  domain.TaskComment  `json:"task_comment"`
	User         domain.User         `json:"user"`
	ActionOrigin domain.ActionOrigin `json:"action_origin,omitempty"`
}

type ProjectCreatedPayload struct {
	Project      domain.Project      `json:"project"`
	User         domain.User         `json:"user"`
	ActionOrigin domain.ActionOrigin `json:"action_origin,omitempty"`
}

type ProjectUpdatedPayload struct {
	Project      domain.Project      `json:"project"`
	User         domain.User         `json:"user"`
	ActionOrigin domain.ActionOrigin `json:"action_origin,omitempty"`
}

type ProjectMemberCreatedPayload struct {
	ProjectMember domain.ProjectMember `json:"project_member"`
	User          domain.User          `json:"user"`
	ActionOrigin  domain.ActionOrigin  `json:"action_origin,omitempty"`
}

type ChatMemberViewedPayload struct {
	ChatMember domain.ChatMember `json:"chat_member"`
	User       domain.User       `json:"user"`
}

type ChatMessageCreatedPayload struct {
	ChatMessage domain.ChatMessage `json:"chat_message"`
	User        domain.User        `json:"user"`
}

type ChatMessageReadPayload struct {
	ChatID    uuid.UUID              `json:"chat_id"`
	MessageID uuid.UUID              `json:"message_id"`
	Read      domain.ChatMessageRead `json:"read"`
}

type ChatMemberCreatedPayload struct {
	ChatMember    domain.ChatMember    `json:"chat_member"`
	ProjectMember domain.ProjectMember `json:"project_member"`
	User          domain.User          `json:"user"`
}

type ChatMemberRemovedPayload struct {
	ChatMember    domain.ChatMember    `json:"chat_member"`
	ProjectMember domain.ProjectMember `json:"project_member"`
	User          domain.User          `json:"user"`
}

type ProjectMemberRemovedPayload struct {
	ProjectMember domain.ProjectMember `json:"project_member"`
	User          domain.User          `json:"user"`
	ActionOrigin  domain.ActionOrigin  `json:"action_origin,omitempty"`
}
