package handlers

import (
	"slices"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	ProjectId   uuid.UUID `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

func (r *CreateTaskRequest) Validate(v *validator.Validator) {
	v.Check("project_id", "project_id is required", r.ProjectId != uuid.Nil)
	v.Check("title", "title is required", validator.NotBlank(r.Title))
	v.Check("description", "description is required", validator.NotBlank(r.Description))
}

type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (r *UpdateTaskRequest) Validate(v *validator.Validator) {
	v.Check("title", "title is required", validator.NotBlank(r.Title))
	v.Check("description", "description is required", validator.NotBlank(r.Description))

	v.Check("status", "status is required", validator.NotBlank(r.Status))

	allowedStatuses := domain.AllowedTaskStatuses
	v.Check("status", "status is invalid", slices.Contains(allowedStatuses, domain.TaskStatus(r.Status)))
}
