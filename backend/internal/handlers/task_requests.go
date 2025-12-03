package handlers

import (
	"slices"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	ProjectId     uuid.UUID  `json:"project_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Priority      string     `json:"priority"`
	ResponsibleId *uuid.UUID `json:"responsible_id"`
	DueDate       *time.Time `json:"due_date"`
	Tags          []string   `json:"tags"`
}

func (r *CreateTaskRequest) Validate(v *validator.Validator) {
	v.Check("project_id", "project_id is required", r.ProjectId != uuid.Nil)
	v.Check("title", "title is required", validator.NotBlank(r.Title))
	v.Check("description", "description is required", validator.NotBlank(r.Description))

	v.Check("priority", "priority is required", validator.NotBlank(r.Priority))
	v.Check("priority", "priority is invalid", slices.Contains(domain.AllowedTaskPriorities, domain.TaskPriority(r.Priority)))

	if r.ResponsibleId != nil && *r.ResponsibleId != uuid.Nil {
		v.Check("responsible_id", "responsible_id is invalid", *r.ResponsibleId != uuid.Nil)
	}

}

type UpdateTaskRequest struct {
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	ResponsibleId *uuid.UUID `json:"responsible_id"`
	DueDate       *time.Time `json:"due_date"`
	Tags          []string   `json:"tags"`
	DoneAt        *time.Time `json:"done_at"`
}

func (r *UpdateTaskRequest) Validate(v *validator.Validator) {
	v.Check("title", "title is required", validator.NotBlank(r.Title))
	v.Check("description", "description is required", validator.NotBlank(r.Description))

	v.Check("status", "status is required", validator.NotBlank(r.Status))

	allowedStatuses := domain.AllowedTaskStatuses
	v.Check("status", "status is invalid", slices.Contains(allowedStatuses, domain.TaskStatus(r.Status)))

	v.Check("priority", "priority is required", validator.NotBlank(r.Priority))
	v.Check("priority", "priority is invalid", slices.Contains(domain.AllowedTaskPriorities, domain.TaskPriority(r.Priority)))

	if r.ResponsibleId != nil && *r.ResponsibleId != uuid.Nil {
		v.Check("responsible_id", "responsible_id is invalid", *r.ResponsibleId != uuid.Nil)
	}

	if r.DueDate != nil {
		v.Check("due_date", "due_date is required", r.DueDate != nil)
	}

}

type MoveTaskRequest struct {
	AfterTaskId *uuid.UUID `json:"after_task_id"`
	ProjectId   uuid.UUID  `json:"project_id"`
	Status      string     `json:"status"`
}

func (r *MoveTaskRequest) Validate(v *validator.Validator) {
	v.Check("status", "status is required", validator.NotBlank(r.Status))
	v.Check("status", "status is invalid", slices.Contains(domain.AllowedTaskStatuses, domain.TaskStatus(r.Status)))

	v.Check("project_id", "project_id is required", r.ProjectId != uuid.Nil)
}
