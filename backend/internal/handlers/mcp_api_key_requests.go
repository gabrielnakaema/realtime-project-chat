package handlers

import (
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/validator"
)

type CreateMCPAPIKeyRequest struct {
	Name   string               `json:"name"`
	Scopes []domain.MCPAPIScope `json:"scopes"`
}

func (req *CreateMCPAPIKeyRequest) Validate(v *validator.Validator) {
	v.Check("name", "name is required", validator.NotBlank(req.Name))
	v.Check("scopes", "at least one scope is required", len(req.Scopes) > 0)
}
