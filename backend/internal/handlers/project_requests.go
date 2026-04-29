package handlers

import (
	"regexp"

	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type ProjectRequest struct {
	Name           string                        `json:"name"`
	Description    string                        `json:"description"`
	Columns        []ProjectColumnRequest        `json:"columns"`
	DeletedColumns []DeletedProjectColumnRequest `json:"deleted_columns"`
}

type ProjectColumnRequest struct {
	Id           *uuid.UUID `json:"id"`
	Name         string     `json:"name"`
	Color        string     `json:"color"`
	IsDoneColumn bool       `json:"is_done_column"`
}

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type DeletedProjectColumnRequest struct {
	Id                  uuid.UUID `json:"id"`
	MoveTasksToColumnId uuid.UUID `json:"move_tasks_to_column_id"`
}

func (r *ProjectRequest) Validate(v *validator.Validator) {
	v.Check("name", "name is required", validator.NotBlank(r.Name))
	v.Check("description", "description is required", validator.NotBlank(r.Description))
	v.Check("columns", "at least one column is required", len(r.Columns) > 0)

	doneColumns := 0
	for i, column := range r.Columns {
		v.Check("columns", "column name is required", validator.NotBlank(column.Name))
		v.Check("columns", "column color must be a valid hex value", hexColorPattern.MatchString(column.Color))
		if column.IsDoneColumn {
			doneColumns++
		}
		if column.Id != nil {
			v.Check("columns", "column id is invalid", *column.Id != uuid.Nil)
		}
		v.Check("columns", "column name is required", validator.NotBlank(r.Columns[i].Name))
	}

	v.Check("columns", "exactly one done column is required", doneColumns == 1)

	for _, deletedColumn := range r.DeletedColumns {
		v.Check("deleted_columns", "deleted column id is invalid", deletedColumn.Id != uuid.Nil)
		v.Check("deleted_columns", "move_tasks_to_column_id is invalid", deletedColumn.MoveTasksToColumnId != uuid.Nil)
	}
}

type CreateMemberRequest struct {
	Email string `json:"email"`
}

func (r *CreateMemberRequest) Validate(v *validator.Validator) {
	v.Check("email", "email is required", validator.NotBlank(r.Email))
	v.Check("email", "email is invalid", validator.ValidEmail(r.Email))
}
