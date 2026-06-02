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
	validateMCPAPIKeyRequest(v, req.Name, req.Scopes)
}

type UpdateMCPAPIKeyRequest struct {
	Name   string               `json:"name"`
	Scopes []domain.MCPAPIScope `json:"scopes"`
}

func (req *UpdateMCPAPIKeyRequest) Validate(v *validator.Validator) {
	validateMCPAPIKeyRequest(v, req.Name, req.Scopes)
}

func validateMCPAPIKeyRequest(v *validator.Validator, name string, scopes []domain.MCPAPIScope) {
	v.Check("name", "name is required", validator.NotBlank(name))
	v.Check("scopes", "at least one scope is required", len(scopes) > 0)
}
