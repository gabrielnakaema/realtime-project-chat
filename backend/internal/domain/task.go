package domain

import (
	"fmt"
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

type Task struct {
	Id          uuid.UUID  `json:"id"`
	ProjectId   uuid.UUID  `json:"project_id"`
	AuthorId    uuid.UUID  `json:"author_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Author  *User        `json:"author,omitempty"`
	Changes []TaskChange `json:"changes,omitempty"`
}

var AllowedTaskStatuses = []TaskStatus{TaskStatusPending, TaskStatusDoing, TaskStatusDone, TaskStatusArchived}

func (t *Task) ChangeStatus(status TaskStatus) error {
	if !slices.Contains(AllowedTaskStatuses, status) {
		return BusinessValidationError("invalid status")
	}

	t.Status = status

	return nil
}

type TaskChange struct {
	Id                uuid.UUID `json:"id"`
	TaskId            uuid.UUID `json:"task_id"`
	AuthorId          uuid.UUID `json:"author_id"`
	ChangeDescription string    `json:"change_description"`
	CreatedAt         time.Time `json:"created_at"`

	Author *User `json:"author,omitempty"`
}

func NewTaskChanges(oldTask *Task, newTask *Task, author *User) []TaskChange {
	changes := []TaskChange{}

	if oldTask.Title != newTask.Title {
		changes = append(changes, TaskChange{
			TaskId:            oldTask.Id,
			AuthorId:          author.Id,
			ChangeDescription: fmt.Sprintf("Title changed from %s to %s by %s", oldTask.Title, newTask.Title, author.Name),
			CreatedAt:         time.Now(),
		})
	}

	if oldTask.Description != newTask.Description {
		changes = append(changes, TaskChange{
			TaskId:            oldTask.Id,
			AuthorId:          author.Id,
			ChangeDescription: fmt.Sprintf("Description changed from %s to %s by %s", oldTask.Description, newTask.Description, author.Name),
			CreatedAt:         time.Now(),
		})
	}

	if oldTask.Status != newTask.Status {
		changes = append(changes, TaskChange{
			TaskId:            oldTask.Id,
			AuthorId:          author.Id,
			ChangeDescription: fmt.Sprintf("Status changed from %s to %s by %s", oldTask.Status, newTask.Status, author.Name),
			CreatedAt:         time.Now(),
		})
	}

	return changes
}
