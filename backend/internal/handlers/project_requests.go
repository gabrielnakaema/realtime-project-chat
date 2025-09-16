package handlers

import (
	"github.com/gabrielnakaema/project-chat/internal/validator"
)

type ProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *ProjectRequest) Validate(v *validator.Validator) {
	v.Check("name", "name is required", validator.NotBlank(r.Name))
	v.Check("description", "description is required", validator.NotBlank(r.Description))
}

type CreateMemberRequest struct {
	Email string `json:"email"`
}

func (r *CreateMemberRequest) Validate(v *validator.Validator) {
	v.Check("email", "email is required", validator.NotBlank(r.Email))
	v.Check("email", "email is invalid", validator.ValidEmail(r.Email))
}
