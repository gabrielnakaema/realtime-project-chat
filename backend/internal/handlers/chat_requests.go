package handlers

import (
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type CreateMessageRequest struct {
	Content string    `json:"content"`
	ChatId  uuid.UUID `json:"chat_id"`
}

func (r *CreateMessageRequest) Validate(v *validator.Validator) {
	v.Check("chat_id", "chat_id is invalid", r.ChatId != uuid.Nil)
	v.Check("content", "content is required", validator.NotBlank(r.Content))
}
