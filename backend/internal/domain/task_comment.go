package domain

import "time"

type TaskComment struct {
	ID              string        `json:"id"`
	Task            *Task         `json:"task"`
	User            *User         `json:"user"`
	Content         string        `json:"content"`
	ParentCommentID *string       `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Replies         []TaskComment `json:"replies"`
}
