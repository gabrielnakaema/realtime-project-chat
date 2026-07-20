package project

import (
	"regexp"

	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type ProjectBody struct {
	Name             string                     `json:"name"`
	Description      string                     `json:"description"`
	RepositoryURL    string                     `json:"repository_url"`
	RepositoryOwner  string                     `json:"repository_owner"`
	RepositoryName   string                     `json:"repository_name"`
	DefaultBranch    string                     `json:"default_branch"`
	BranchNamePrefix string                     `json:"branch_name_prefix"`
	Columns          []ProjectColumnBody        `json:"columns"`
	DeletedColumns   []DeletedProjectColumnBody `json:"deleted_columns"`
}

type ProjectColumnBody struct {
	Id           *uuid.UUID `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Color        string     `json:"color"`
	IsDoneColumn bool       `json:"is_done_column"`
}

type UpdateProjectColumnBody struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	IsDoneColumn bool   `json:"is_done_column"`
}

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type DeletedProjectColumnBody struct {
	Id                  uuid.UUID `json:"id"`
	MoveTasksToColumnId uuid.UUID `json:"move_tasks_to_column_id"`
}

func (r *ProjectBody) Validate(v *validator.Validator) {
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

func (r *UpdateProjectColumnBody) Validate(v *validator.Validator) {
	v.Check("name", "name is required", validator.NotBlank(r.Name))
	v.Check("color", "color must be a valid hex value", hexColorPattern.MatchString(r.Color))
}

type CreateMemberBody struct {
	Email string `json:"email"`
}

func (r *CreateMemberBody) Validate(v *validator.Validator) {
	v.Check("email", "email is required", validator.NotBlank(r.Email))
	v.Check("email", "email is invalid", validator.ValidEmail(r.Email))
}
