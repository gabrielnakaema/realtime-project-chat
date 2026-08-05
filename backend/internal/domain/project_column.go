package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProjectColumnNotFound = errors.New("invalid project_column_id")
	ErrDoneColumnNotUnique   = errors.New("project must have exactly one done column")
)

func (p *Project) Column(id uuid.UUID) (*ProjectColumn, error) {
	for _, column := range p.Columns {
		if column.Id == id {
			return &column, nil
		}
	}

	return nil, ErrProjectColumnNotFound
}

func (p *Project) ValidateColumnIDs(ids []uuid.UUID) error {
	for _, id := range ids {
		if _, err := p.Column(id); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) DoneColumn() (*ProjectColumn, error) {
	var done *ProjectColumn
	for _, column := range p.Columns {
		if !column.IsDoneColumn {
			continue
		}
		if done != nil {
			return nil, ErrDoneColumnNotUnique
		}
		done = &column
	}

	if done == nil {
		return nil, ErrDoneColumnNotUnique
	}

	return done, nil
}

// legacy status field for compatibility with previous versions
func TaskStatusIn(column *ProjectColumn, archivedAt *time.Time) TaskStatus {
	if archivedAt != nil {
		return TaskStatusArchived
	}

	if column == nil {
		return ""
	}

	return TaskStatus(strings.ToLower(column.Name))
}

func (c *ProjectColumn) CompletedAt(current *time.Time) *time.Time {
	if !c.IsDoneColumn {
		return nil
	}

	if current != nil {
		return current
	}

	now := time.Now()
	return &now
}
