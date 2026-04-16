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

type GetOrCreateGeneralChatRequest struct {
	UserIds []uuid.UUID `json:"user_ids"`
}

func (r *GetOrCreateGeneralChatRequest) Validate(v *validator.Validator) {
	v.Check("user_ids", "at least one user is required", len(r.UserIds) > 0)

	for _, userId := range r.UserIds {
		v.Check("user_ids", "user_ids contains an invalid user", userId != uuid.Nil)
	}
}
