package events

import (
	"encoding/json"

	"github.com/gabrielnakaema/project-chat/internal/domain"
)

type Payload interface {
	ToMessage() ([]byte, error)
}

type TaskCreatedPayload struct {
	Task domain.Task `json:"task"`
	User domain.User `json:"user"`
}

func (t *TaskCreatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(t)
}

type TaskUpdatedPayload struct {
	Task           domain.Task        `json:"task"`
	PreviousTask   *domain.Task       `json:"previous_task,omitempty"`
	User           domain.User        `json:"user"`
	PreviousStatus *domain.TaskStatus `json:"previous_status,omitempty"`
}

func (t *TaskUpdatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(t)
}

type ProjectCreatedPayload struct {
	Project domain.Project `json:"project"`
	User    domain.User    `json:"user"`
}

func (p *ProjectCreatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(p)
}

type ProjectUpdatedPayload struct {
	Project domain.Project `json:"project"`
	User    domain.User    `json:"user"`
}

func (p *ProjectUpdatedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(p)
}

type ProjectMemberCreatedPayload struct {
	ProjectMember domain.ProjectMember `json:"project_member"`
	User          domain.User          `json:"user"`
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
}

func (p *ProjectMemberRemovedPayload) ToMessage() ([]byte, error) {
	return json.Marshal(p)
}
