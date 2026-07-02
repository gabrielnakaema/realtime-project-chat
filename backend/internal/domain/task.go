package domain

import (
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TaskDependencyRef struct {
	Id    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Code  string    `json:"code"`
}

type TaskCodeSuggestion struct {
	Code string `json:"code"`
	Kind string `json:"kind"`
}

type TaskDependencyEdge struct {
	TaskId          uuid.UUID
	DependsOnTaskId uuid.UUID
}

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
	Id               uuid.UUID           `json:"id"`
	ProjectId        uuid.UUID           `json:"project_id"`
	AuthorId         uuid.UUID           `json:"author_id"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	Code             string              `json:"code"`
	Status           TaskStatus          `json:"status"`
	ProjectColumnId  uuid.UUID           `json:"project_column_id"`
	Priority         TaskPriority        `json:"priority"`
	Order            string              `json:"order"`
	ResponsibleId    *uuid.UUID          `json:"responsible_id"`
	DueDate          *time.Time          `json:"due_date"`
	DoneAt           *time.Time          `json:"done_at"`
	ArchivedAt       *time.Time          `json:"archived_at"`
	Tags             []string            `json:"tags"`
	DependsOnTaskIds []uuid.UUID         `json:"depends_on_task_ids"`
	DependsOnTasks   []TaskDependencyRef `json:"depends_on_tasks,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`

	Responsible   *User          `json:"responsible,omitempty"`
	Author        *User          `json:"author,omitempty"`
	Updates       []TaskUpdate   `json:"updates,omitempty"`
	Project       *Project       `json:"project,omitempty"`
	ProjectColumn *ProjectColumn `json:"project_column,omitempty"`
	Comments      []TaskComment  `json:"comments,omitempty"`
}

type TaskUpdateType string

var (
	TaskUpdateTypeCreated    TaskUpdateType = "created"
	TaskUpdateTypeColumn     TaskUpdateType = "column"
	TaskUpdateTypeUpdated    TaskUpdateType = "updated"
	TaskUpdateTypeDone       TaskUpdateType = "done"
	TaskUpdateTypeArchived   TaskUpdateType = "archived"
	TaskUpdateTypeAssigned   TaskUpdateType = "assigned"
	TaskUpdateTypeUnassigned TaskUpdateType = "unassigned"
)

type TaskUpdate struct {
	Id           uuid.UUID      `json:"id"`
	TaskId       uuid.UUID      `json:"task_id"`
	UserId       uuid.UUID      `json:"user_id"`
	UpdateType   TaskUpdateType `json:"update_type"`
	ActionOrigin ActionOrigin   `json:"action_origin"`
	CreatedAt    time.Time      `json:"created_at"`

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
		TaskId:       task.Id,
		UserId:       author.Id,
		UpdateType:   TaskUpdateTypeCreated,
		ActionOrigin: ActionOriginUser,
		Changes:      []TaskChange{},
		CreatedAt:    time.Now(),
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
		field: "code",
		build: func(old *Task, new *Task) *TaskChange {
			if old.Code == new.Code {
				return nil
			}

			return &TaskChange{
				Field:    "code",
				OldValue: old.Code,
				NewValue: new.Code,
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
		field: "column",
		build: func(old *Task, new *Task) *TaskChange {
			if old.ProjectColumnId == new.ProjectColumnId {
				return nil
			}

			oldDisplay := ""
			if old.ProjectColumn != nil {
				oldDisplay = old.ProjectColumn.Name
			}

			newDisplay := ""
			if new.ProjectColumn != nil {
				newDisplay = new.ProjectColumn.Name
			}

			return &TaskChange{
				Field:           "column",
				OldValue:        old.ProjectColumnId.String(),
				NewValue:        new.ProjectColumnId.String(),
				OldDisplayValue: optionalString(oldDisplay),
				NewDisplayValue: optionalString(newDisplay),
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
	{
		field: "depends_on_task_ids",
		build: func(old *Task, new *Task) *TaskChange {
			oldValue := joinSortedUUIDs(old.DependsOnTaskIds)
			newValue := joinSortedUUIDs(new.DependsOnTaskIds)
			if oldValue == newValue {
				return nil
			}

			oldDisplay := formatSortedRefs(old.DependsOnTasks)
			newDisplay := formatSortedRefs(new.DependsOnTasks)
			return &TaskChange{
				Field:           "depends_on_task_ids",
				OldValue:        oldValue,
				NewValue:        newValue,
				OldDisplayValue: optionalString(oldDisplay),
				NewDisplayValue: optionalString(newDisplay),
			}
		},
	},
}

func NewTaskUpdate(old *Task, new *Task, author *User) TaskUpdate {
	changes := buildTaskChanges(old, new)

	return TaskUpdate{
		TaskId:       old.Id,
		UserId:       author.Id,
		UpdateType:   determineTaskUpdateType(changes),
		ActionOrigin: ActionOriginUser,
		CreatedAt:    time.Now(),
		Changes:      changes,
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

	if len(changes) == 1 && changes[0].Field == "column" {
		return TaskUpdateTypeColumn
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

func TaskDependencyRefsToIDs(refs []TaskDependencyRef) []uuid.UUID {
	ids := make([]uuid.UUID, len(refs))
	for i, ref := range refs {
		ids[i] = ref.Id
	}
	return ids
}

func joinSortedUUIDs(ids []uuid.UUID) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.String()
	}
	sort.Strings(strs)
	return strings.Join(strs, ",")
}

func formatSortedRefs(refs []TaskDependencyRef) string {
	if len(refs) == 0 {
		return ""
	}

	type sortable struct {
		id    string
		label string
	}
	items := make([]sortable, len(refs))
	for i, ref := range refs {
		label := ref.Title
		if ref.Code != "" {
			label = ref.Code + " — " + ref.Title
		}
		items[i] = sortable{id: ref.Id.String(), label: label}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].id < items[j].id })

	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.label
	}
	return strings.Join(labels, ", ")
}

func stringPointer(value string) *string {
	return &value
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
