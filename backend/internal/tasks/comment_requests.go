package tasks

import (
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type CreateTaskCommentBody struct {
	Content         string     `json:"content"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id"`
}

func (r *CreateTaskCommentBody) Validate(v *validator.Validator) {
	v.Check("content", "content is required", validator.NotBlank(r.Content))

	if r.ParentCommentID != nil {
		v.Check("parent_comment_id", "parent_comment_id is invalid", *r.ParentCommentID != uuid.Nil)
	}
}
