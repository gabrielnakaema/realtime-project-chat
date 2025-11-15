package domain

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

var (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusDoing    TaskStatus = "doing"
	TaskStatusDone     TaskStatus = "done"
	TaskStatusArchived TaskStatus = "archived"
)

type TaskPriority string

var (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

var AllowedTaskPriorities = []TaskPriority{TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh}

type Task struct {
	Id            uuid.UUID    `json:"id"`
	ProjectId     uuid.UUID    `json:"project_id"`
	AuthorId      uuid.UUID    `json:"author_id"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	Status        TaskStatus   `json:"status"`
	Priority      TaskPriority `json:"priority"`
	Order         int          `json:"order"`
	ResponsibleId *uuid.UUID   `json:"responsible_id"`
	DueDate       *time.Time   `json:"due_date"`
	DoneAt        *time.Time   `json:"done_at"`
	Tags          []string     `json:"tags"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`

	Responsible *User        `json:"responsible,omitempty"`
	Author      *User        `json:"author,omitempty"`
	Updates     []TaskUpdate `json:"updates,omitempty"`
}

var AllowedTaskStatuses = []TaskStatus{TaskStatusPending, TaskStatusDoing, TaskStatusDone, TaskStatusArchived}

func (t *Task) ChangeStatus(status TaskStatus) error {
	if !slices.Contains(AllowedTaskStatuses, status) {
		return BusinessValidationError("invalid status")
	}

	t.Status = status

	return nil
}

type TaskUpdateType string

var (
	TaskUpdateTypeCreated    TaskUpdateType = "created"
	TaskUpdateTypeStatus     TaskUpdateType = "status"
	TaskUpdateTypeUpdated    TaskUpdateType = "updated"
	TaskUpdateTypeDone       TaskUpdateType = "done"
	TaskUpdateTypeArchived   TaskUpdateType = "archived"
	TaskUpdateTypeAssigned   TaskUpdateType = "assigned"
	TaskUpdateTypeUnassigned TaskUpdateType = "unassigned"
)

type TaskUpdate struct {
	Id         uuid.UUID      `json:"id"`
	TaskId     uuid.UUID      `json:"task_id"`
	UserId     uuid.UUID      `json:"user_id"`
	UpdateType TaskUpdateType `json:"update_type"`
	CreatedAt  time.Time      `json:"created_at"`

	User    *User        `json:"user,omitempty"`
	Changes []TaskChange `json:"changes,omitempty"`
}

type TaskChange struct {
	Id        uuid.UUID  `json:"id"`
	UpdateId  uuid.UUID  `json:"update_id"`
	SubjectId *uuid.UUID `json:"subject_id"`
	Field     string     `json:"field"`
	OldValue  string     `json:"old_value"`
	NewValue  string     `json:"new_value"`
	CreatedAt time.Time  `json:"created_at"`

	Subject *User `json:"subject,omitempty"`
}

func NewTaskCreatedUpdate(task *Task, author *User) TaskUpdate {
	return TaskUpdate{
		TaskId:     task.Id,
		UserId:     author.Id,
		UpdateType: TaskUpdateTypeCreated,
		Changes:    []TaskChange{},
		CreatedAt:  time.Now(),
	}
}

func NewTaskUpdate(old *Task, new *Task, author *User) TaskUpdate {
	changes := []TaskChange{}

	if old.Title != new.Title {
		changes = append(changes, TaskChange{
			Field:    "title",
			OldValue: old.Title,
			NewValue: new.Title,
		})
	}

	if old.Description != new.Description {
		changes = append(changes, TaskChange{
			Field:    "description",
			OldValue: old.Description,
			NewValue: new.Description,
		})
	}

	if old.Status != new.Status {
		changes = append(changes, TaskChange{
			Field:    "status",
			OldValue: string(old.Status),
			NewValue: string(new.Status),
		})
	}

	if old.Priority != new.Priority {
		changes = append(changes, TaskChange{
			Field:    "priority",
			OldValue: string(old.Priority),
			NewValue: string(new.Priority),
		})
	}

	if old.ResponsibleId != new.ResponsibleId {
		oldResponsibleId := ""
		if old.ResponsibleId != nil {
			oldResponsibleId = old.ResponsibleId.String()
		}
		newResponsibleId := ""
		if new.ResponsibleId != nil {
			newResponsibleId = new.ResponsibleId.String()
		}

		changes = append(changes, TaskChange{
			Field:    "responsible_id",
			OldValue: oldResponsibleId,
			NewValue: newResponsibleId,
		})
	}

	if old.DueDate != new.DueDate {
		oldDueDate := ""
		if old.DueDate != nil {
			oldDueDate = old.DueDate.Format(time.RFC3339)
		}
		newDueDate := ""
		if new.DueDate != nil {
			newDueDate = new.DueDate.Format(time.RFC3339)
		}

		changes = append(changes, TaskChange{
			Field:    "due_date",
			OldValue: oldDueDate,
			NewValue: newDueDate,
		})
	}

	if old.DoneAt != new.DoneAt {
		oldDoneAt := ""
		if old.DoneAt != nil {
			oldDoneAt = old.DoneAt.Format(time.RFC3339)
		}
		newDoneAt := ""
		if new.DoneAt != nil {
			newDoneAt = new.DoneAt.Format(time.RFC3339)
		}

		changes = append(changes, TaskChange{
			Field:    "done_at",
			OldValue: oldDoneAt,
			NewValue: newDoneAt,
		})
	}

	return TaskUpdate{
		TaskId:     old.Id,
		UserId:     author.Id,
		UpdateType: TaskUpdateTypeUpdated,
		CreatedAt:  time.Now(),
		Changes:    changes,
	}
}
