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
	Order         string       `json:"order"`
	ResponsibleId *uuid.UUID   `json:"responsible_id"`
	DueDate       *time.Time   `json:"due_date"`
	DoneAt        *time.Time   `json:"done_at"`
	Tags          []string     `json:"tags"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`

	Responsible *User        `json:"responsible,omitempty"`
	Author      *User        `json:"author,omitempty"`
	Updates     []TaskUpdate `json:"updates,omitempty"`
	Project     *Project     `json:"project,omitempty"`
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
	Id              uuid.UUID  `json:"id"`
	UpdateId        uuid.UUID  `json:"update_id"`
	SubjectId       *uuid.UUID `json:"subject_id"`
	OldValueId      *uuid.UUID `json:"old_value_id"`
	NewValueId      *uuid.UUID `json:"new_value_id"`
	Field           string     `json:"field"`
	OldValue        string     `json:"old_value"`
	NewValue        string     `json:"new_value"`
	OldDisplayValue *string    `json:"old_display_value,omitempty"`
	NewDisplayValue *string    `json:"new_display_value,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`

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

type taskChangeDefinition struct {
	field string
	build func(old *Task, new *Task) *TaskChange
}

var taskChangeDefinitions = []taskChangeDefinition{
	{
		field: "title",
		build: func(old *Task, new *Task) *TaskChange {
			if old.Title == new.Title {
				return nil
			}

			return &TaskChange{
				Field:    "title",
				OldValue: old.Title,
				NewValue: new.Title,
			}
		},
	},
	{
		field: "description",
		build: func(old *Task, new *Task) *TaskChange {
			if old.Description == new.Description {
				return nil
			}

			return &TaskChange{
				Field:    "description",
				OldValue: old.Description,
				NewValue: new.Description,
			}
		},
	},
	{
		field: "status",
		build: func(old *Task, new *Task) *TaskChange {
			if old.Status == new.Status {
				return nil
			}

			return &TaskChange{
				Field:    "status",
				OldValue: string(old.Status),
				NewValue: string(new.Status),
			}
		},
	},
	{
		field: "priority",
		build: func(old *Task, new *Task) *TaskChange {
			if old.Priority == new.Priority {
				return nil
			}

			return &TaskChange{
				Field:    "priority",
				OldValue: string(old.Priority),
				NewValue: string(new.Priority),
			}
		},
	},
	{
		field: "responsible_id",
		build: buildResponsibleChange,
	},
	{
		field: "due_date",
		build: func(old *Task, new *Task) *TaskChange {
			oldDueDate := formatTaskTime(old.DueDate)
			newDueDate := formatTaskTime(new.DueDate)
			if oldDueDate == newDueDate {
				return nil
			}

			return &TaskChange{
				Field:    "due_date",
				OldValue: oldDueDate,
				NewValue: newDueDate,
			}
		},
	},
	{
		field: "done_at",
		build: func(old *Task, new *Task) *TaskChange {
			oldDoneAt := formatTaskTime(old.DoneAt)
			newDoneAt := formatTaskTime(new.DoneAt)
			if oldDoneAt == newDoneAt {
				return nil
			}

			return &TaskChange{
				Field:    "done_at",
				OldValue: oldDoneAt,
				NewValue: newDoneAt,
			}
		},
	},
}

func NewTaskUpdate(old *Task, new *Task, author *User) TaskUpdate {
	changes := buildTaskChanges(old, new)

	return TaskUpdate{
		TaskId:     old.Id,
		UserId:     author.Id,
		UpdateType: determineTaskUpdateType(changes),
		CreatedAt:  time.Now(),
		Changes:    changes,
	}
}

func buildTaskChanges(old *Task, new *Task) []TaskChange {
	changes := []TaskChange{}

	for _, definition := range taskChangeDefinitions {
		change := definition.build(old, new)
		if change == nil {
			continue
		}

		changes = append(changes, *change)
	}

	return changes
}

func determineTaskUpdateType(changes []TaskChange) TaskUpdateType {
	if len(changes) == 1 && changes[0].Field == "responsible_id" {
		switch {
		case changes[0].OldValue == "" && changes[0].NewValue != "":
			return TaskUpdateTypeAssigned
		case changes[0].OldValue != "" && changes[0].NewValue == "":
			return TaskUpdateTypeUnassigned
		}
	}

	if len(changes) == 1 && changes[0].Field == "status" {
		return TaskUpdateTypeStatus
	}

	return TaskUpdateTypeUpdated
}

func buildResponsibleChange(old *Task, new *Task) *TaskChange {
	oldRaw, oldID, oldDisplay := taskRelationValues(old.ResponsibleId, old.Responsible)
	newRaw, newID, newDisplay := taskRelationValues(new.ResponsibleId, new.Responsible)

	if oldRaw == newRaw {
		return nil
	}

	change := &TaskChange{
		Field:           "responsible_id",
		OldValue:        oldRaw,
		NewValue:        newRaw,
		OldValueId:      oldID,
		NewValueId:      newID,
		OldDisplayValue: oldDisplay,
		NewDisplayValue: newDisplay,
	}

	switch {
	case oldID == nil && newID != nil:
		change.SubjectId = newID
	case oldID != nil && newID == nil:
		change.SubjectId = oldID
	}

	return change
}

func taskRelationValues(id *uuid.UUID, user *User) (string, *uuid.UUID, *string) {
	if id == nil {
		return "", nil, nil
	}

	raw := id.String()
	var display *string
	if user != nil && user.Name != "" {
		display = stringPointer(user.Name)
	}

	idCopy := *id
	return raw, &idCopy, display
}

func formatTaskTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.UTC().Format(time.RFC3339)
}

func stringPointer(value string) *string {
	return &value
}
