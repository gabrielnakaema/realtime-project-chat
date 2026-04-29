package handlers

import (
	"slices"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	ProjectId       uuid.UUID  `json:"project_id"`
	ProjectColumnId uuid.UUID  `json:"project_column_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Priority        string     `json:"priority"`
	ResponsibleId   *uuid.UUID `json:"responsible_id"`
	DueDate         *time.Time `json:"due_date"`
	Tags            []string   `json:"tags"`
}

func (r *CreateTaskRequest) Validate(v *validator.Validator) {
	v.Check("project_id", "project_id is required", r.ProjectId != uuid.Nil)
	v.Check("project_column_id", "project_column_id is required", r.ProjectColumnId != uuid.Nil)
	v.Check("title", "title is required", validator.NotBlank(r.Title))
	v.Check("description", "description is required", validator.NotBlank(r.Description))

	v.Check("priority", "priority is required", validator.NotBlank(r.Priority))
	v.Check("priority", "priority is invalid", slices.Contains(domain.AllowedTaskPriorities, domain.TaskPriority(r.Priority)))

	if r.ResponsibleId != nil && *r.ResponsibleId != uuid.Nil {
		v.Check("responsible_id", "responsible_id is invalid", *r.ResponsibleId != uuid.Nil)
	}
}

type UpdateTaskRequest struct {
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	ProjectColumnId uuid.UUID  `json:"project_column_id"`
	Priority        string     `json:"priority"`
	ResponsibleId   *uuid.UUID `json:"responsible_id"`
	DueDate         *time.Time `json:"due_date"`
	Tags            []string   `json:"tags"`
}

func (r *UpdateTaskRequest) Validate(v *validator.Validator) {
	v.Check("title", "title is required", validator.NotBlank(r.Title))
	v.Check("description", "description is required", validator.NotBlank(r.Description))

	v.Check("project_column_id", "project_column_id is required", r.ProjectColumnId != uuid.Nil)

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
	AfterTaskId     *uuid.UUID `json:"after_task_id"`
	ProjectId       uuid.UUID  `json:"project_id"`
	ProjectColumnId uuid.UUID  `json:"project_column_id"`
}

func (r *MoveTaskRequest) Validate(v *validator.Validator) {
	v.Check("project_id", "project_id is required", r.ProjectId != uuid.Nil)
	v.Check("project_column_id", "project_column_id is required", r.ProjectColumnId != uuid.Nil)
}
