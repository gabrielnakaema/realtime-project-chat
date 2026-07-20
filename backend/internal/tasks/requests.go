package tasks

import (
	"slices"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type CreateTaskBody struct {
	ProjectId        uuid.UUID   `json:"project_id"`
	ProjectColumnId  uuid.UUID   `json:"project_column_id"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Code             string      `json:"code"`
	Priority         string      `json:"priority"`
	ResponsibleId    *uuid.UUID  `json:"responsible_id"`
	DueDate          *time.Time  `json:"due_date"`
	Tags             []string    `json:"tags"`
	DependsOnTaskIds []uuid.UUID `json:"depends_on_task_ids"`
}

func (r *CreateTaskBody) Validate(v *validator.Validator) {
	v.Check("project_id", "project_id is required", r.ProjectId != uuid.Nil)
	v.Check("project_column_id", "project_column_id is required", r.ProjectColumnId != uuid.Nil)
	v.Check("title", "title is required", validator.NotBlank(r.Title))
	v.Check("description", "description is required", validator.NotBlank(r.Description))

	v.Check("priority", "priority is required", validator.NotBlank(r.Priority))
	v.Check("priority", "priority is invalid", slices.Contains(domain.AllowedTaskPriorities, domain.TaskPriority(r.Priority)))

}

type UpdateTaskBody struct {
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Code             string      `json:"code"`
	ProjectColumnId  uuid.UUID   `json:"project_column_id"`
	Priority         string      `json:"priority"`
	ResponsibleId    *uuid.UUID  `json:"responsible_id"`
	DueDate          *time.Time  `json:"due_date"`
	Tags             []string    `json:"tags"`
	DependsOnTaskIds []uuid.UUID `json:"depends_on_task_ids"`
}

func (r *UpdateTaskBody) Validate(v *validator.Validator) {
	v.Check("title", "title is required", validator.NotBlank(r.Title))
	v.Check("description", "description is required", validator.NotBlank(r.Description))

	v.Check("project_column_id", "project_column_id is required", r.ProjectColumnId != uuid.Nil)

	v.Check("priority", "priority is required", validator.NotBlank(r.Priority))
	v.Check("priority", "priority is invalid", slices.Contains(domain.AllowedTaskPriorities, domain.TaskPriority(r.Priority)))

}

type MoveTaskBody struct {
	AfterTaskId     *uuid.UUID `json:"after_task_id"`
	ProjectId       uuid.UUID  `json:"project_id"`
	ProjectColumnId uuid.UUID  `json:"project_column_id"`
}

func (r *MoveTaskBody) Validate(v *validator.Validator) {
	v.Check("project_id", "project_id is required", r.ProjectId != uuid.Nil)
	v.Check("project_column_id", "project_column_id is required", r.ProjectColumnId != uuid.Nil)
}

type RestoreTaskBody struct {
	ProjectColumnId uuid.UUID `json:"project_column_id"`
}

func (r *RestoreTaskBody) Validate(v *validator.Validator) {
	v.Check("project_column_id", "project_column_id is required", r.ProjectColumnId != uuid.Nil)
}
