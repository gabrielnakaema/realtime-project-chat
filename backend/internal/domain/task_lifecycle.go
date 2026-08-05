package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotArchived     = errors.New("task is not archived")
	ErrTaskAlreadyArchived = errors.New("task is already archived")
	ErrTaskArchived        = errors.New("task is archived")
)

type TaskEdit struct {
	Title          string
	Description    string
	Code           *string
	Priority       TaskPriority
	DueDate        *time.Time
	ResponsibleId  *uuid.UUID
	Responsible    *User
	Tags           []string
	DependsOnTasks []TaskDependencyRef
}

type TaskPlacement struct {
	TaskId          uuid.UUID
	ProjectId       uuid.UUID
	ProjectColumnId uuid.UUID
	Order           string
	DoneAt          *time.Time
}

func (t *Task) IsArchived() bool {
	return t.ArchivedAt != nil
}

func (t *Task) Archive() (*Task, error) {
	if t.IsArchived() {
		return nil, ErrTaskAlreadyArchived
	}

	now := time.Now()

	archived := *t
	archived.ArchivedAt = &now
	archived.Status = TaskStatusIn(t.ProjectColumn, &now)
	archived.UpdatedAt = now
	archived.Updates = []TaskUpdate{}

	return &archived, nil
}

func (t *Task) Restore(project *Project, columnID uuid.UUID) (*Task, error) {
	if !t.IsArchived() {
		return nil, ErrTaskNotArchived
	}

	column, err := project.Column(columnID)
	if err != nil {
		return nil, err
	}

	restored := *t
	restored.ProjectColumnId = column.Id
	restored.ProjectColumn = column
	restored.ArchivedAt = nil
	restored.Status = TaskStatusIn(column, nil)
	restored.DoneAt = column.CompletedAt(nil)
	restored.UpdatedAt = time.Now()
	restored.Updates = []TaskUpdate{}

	return &restored, nil
}

func (t *Task) MoveTo(column *ProjectColumn, order string) *TaskPlacement {
	return &TaskPlacement{
		TaskId:          t.Id,
		ProjectId:       t.ProjectId,
		Order:           order,
		ProjectColumnId: column.Id,
		DoneAt:          column.CompletedAt(t.DoneAt),
	}
}

func (t *Task) ApplyEdit(edit TaskEdit, column *ProjectColumn) *Task {
	code := t.Code
	if edit.Code != nil {
		code = strings.TrimSpace(*edit.Code)
	}

	return &Task{
		Id:               t.Id,
		ProjectId:        t.ProjectId,
		ProjectColumnId:  column.Id,
		AuthorId:         t.AuthorId,
		CreatedAt:        t.CreatedAt,
		Author:           t.Author,
		Order:            t.Order,
		Title:            edit.Title,
		Description:      edit.Description,
		Code:             code,
		Status:           TaskStatusIn(column, t.ArchivedAt),
		Priority:         edit.Priority,
		ResponsibleId:    edit.ResponsibleId,
		Responsible:      edit.Responsible,
		DueDate:          edit.DueDate,
		Tags:             NormalizeTaskTags(edit.Tags),
		DependsOnTaskIds: TaskDependencyRefsToIDs(edit.DependsOnTasks),
		DependsOnTasks:   edit.DependsOnTasks,
		DoneAt:           column.CompletedAt(t.DoneAt),
		UpdatedAt:        time.Now(),
		Updates:          []TaskUpdate{},
		ProjectColumn:    column,
		ArchivedAt:       t.ArchivedAt,
	}
}

func (t *Task) IsAssignedTo(userID uuid.UUID) bool {
	return t.ResponsibleId != nil && *t.ResponsibleId == userID
}

func (t *Task) IsCompletedIn(column *ProjectColumn) bool {
	return t.ProjectColumnId == column.Id && t.DoneAt != nil
}

func NormalizeTaskTags(tags []string) []string {
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		if strings.TrimSpace(tag) == "" {
			continue
		}

		normalized = append(normalized, tag)
	}

	return normalized
}
