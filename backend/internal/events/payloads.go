package events

import (
	"encoding/json"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
)

type Payload interface {
	ToMessage() ([]byte, error)
}

type TaskCreatedPayload struct {
	Task         domain.Task         `json:"task"`
	User         domain.User         `json:"user"`
	ActionOrigin domain.ActionOrigin `json:"action_origin,omitempty"`
}

func (t *TaskCreatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(t)
}

type TaskUpdatedPayload struct {
	Task                    domain.Task         `json:"task"`
	PreviousTask            *domain.Task        `json:"previous_task,omitempty"`
	User                    domain.User         `json:"user"`
	PreviousProjectColumnID *uuid.UUID          `json:"previous_project_column_id,omitempty"`
	ActionOrigin            domain.ActionOrigin `json:"action_origin,omitempty"`
}

func (t *TaskUpdatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(t)
}

type TaskCommentCreatedPayload struct {
	TaskComment  domain.TaskComment  `json:"task_comment"`
	User         domain.User         `json:"user"`
	ActionOrigin domain.ActionOrigin `json:"action_origin,omitempty"`
}

func (t *TaskCommentCreatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(t)
}

type ProjectCreatedPayload struct {
	Project      domain.Project      `json:"project"`
	User         domain.User         `json:"user"`
	ActionOrigin domain.ActionOrigin `json:"action_origin,omitempty"`
}

func (p *ProjectCreatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(p)
}

type ProjectUpdatedPayload struct {
	Project      domain.Project      `json:"project"`
	User         domain.User         `json:"user"`
	ActionOrigin domain.ActionOrigin `json:"action_origin,omitempty"`
}

func (p *ProjectUpdatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(p)
}

type ProjectMemberCreatedPayload struct {
	ProjectMember domain.ProjectMember `json:"project_member"`
	User          domain.User          `json:"user"`
	ActionOrigin  domain.ActionOrigin  `json:"action_origin,omitempty"`
}

func (p *ProjectMemberCreatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(p)
}

type ChatMemberViewedPayload struct {
	ChatMember domain.ChatMember `json:"chat_member"`
	User       domain.User       `json:"user"`
}

func (c *ChatMemberViewedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(c)
}

type ChatMessageCreatedPayload struct {
	ChatMessage domain.ChatMessage `json:"chat_message"`
	User        domain.User        `json:"user"`
}

func (c *ChatMessageCreatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(c)
}

type ChatMessageReadPayload struct {
	ChatID    uuid.UUID              `json:"chat_id"`
	MessageID uuid.UUID              `json:"message_id"`
	Read      domain.ChatMessageRead `json:"read"`
}

func (c *ChatMessageReadPayload) ToMessage() ([]byte, error) {
	return json.Marshal(c)
}

type ChatMemberCreatedPayload struct {
	ChatMember domain.ChatMember `json:"chat_member"`
	User       domain.User       `json:"user"`
}

func (c *ChatMemberCreatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(c)
}

type ProjectMemberRemovedPayload struct {
	ProjectMember domain.ProjectMember `json:"project_member"`
	User          domain.User          `json:"user"`
	ActionOrigin  domain.ActionOrigin  `json:"action_origin,omitempty"`
}

func (p *ProjectMemberRemovedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(p)
}
