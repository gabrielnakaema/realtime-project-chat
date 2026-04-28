package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeProjectMemberCreated NotificationType = "project.member.created"
	NotificationTypeTaskAssigned         NotificationType = "task.assigned"
	NotificationTypeTaskCommentCreated   NotificationType = "task.comment.created"
)

type Notification struct {
	Id            uuid.UUID        `json:"id"`
	UserId        uuid.UUID        `json:"user_id"`
	ActorId       uuid.UUID        `json:"actor_id"`
	ProjectId     uuid.UUID        `json:"project_id"`
	TaskId        *uuid.UUID       `json:"task_id,omitempty"`
	TaskCommentId *uuid.UUID       `json:"task_comment_id,omitempty"`
	Type          NotificationType `json:"type"`
	ReadAt        *time.Time       `json:"read_at,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`

	Actor       *User        `json:"actor,omitempty"`
	Project     *Project     `json:"project,omitempty"`
	Task        *Task        `json:"task,omitempty"`
	TaskComment *TaskComment `json:"task_comment,omitempty"`
}
