package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskArchive(t *testing.T) {
	updatedAt := time.Now().Add(-time.Hour)

	newTask := func() *Task {
		return &Task{
			Id:              uuid.New(),
			ProjectId:       uuid.New(),
			Title:           "Task",
			Status:          TaskStatusPending,
			ProjectColumnId: uuid.New(),
			ProjectColumn:   &ProjectColumn{Name: "Pending"},
			Version:         7,
			UpdatedAt:       updatedAt,
			Updates:         []TaskUpdate{{UpdateType: TaskUpdateTypeCreated}},
		}
	}

	task := newTask()
	archived, err := task.Archive()
	require.NoError(t, err)

	t.Run("takes the task off the board", func(t *testing.T) {
		assert.Equal(t, TaskStatusArchived, archived.Status)
		require.NotNil(t, archived.ArchivedAt)
		assert.WithinDuration(t, time.Now(), *archived.ArchivedAt, time.Minute)
		assert.True(t, archived.UpdatedAt.After(updatedAt))
	})

	t.Run("keeps the rest of the task", func(t *testing.T) {
		assert.Equal(t, task.Id, archived.Id)
		assert.Equal(t, task.ProjectId, archived.ProjectId)
		assert.Equal(t, task.Title, archived.Title)
		assert.Equal(t, task.ProjectColumnId, archived.ProjectColumnId)
	})

	t.Run("leaves the version to the database", func(t *testing.T) {
		assert.Equal(t, 7, archived.Version)
	})

	t.Run("drops the update history the caller loaded", func(t *testing.T) {
		assert.Empty(t, archived.Updates)
	})

	t.Run("does not touch the task it was called on", func(t *testing.T) {
		assert.Equal(t, TaskStatusPending, task.Status)
		assert.Nil(t, task.ArchivedAt)
		assert.Equal(t, updatedAt, task.UpdatedAt)
		assert.Len(t, task.Updates, 1)
	})

	t.Run("keeps the completion of a task archived after it was done", func(t *testing.T) {
		doneAt := time.Now().Add(-2 * time.Hour)
		done := newTask()
		done.ProjectColumn = &ProjectColumn{Name: "Done", IsDoneColumn: true}
		done.Status = TaskStatusDone
		done.DoneAt = &doneAt

		archived, err := done.Archive()

		require.NoError(t, err)
		require.NotNil(t, archived.DoneAt)
		assert.Equal(t, doneAt, *archived.DoneAt)
	})

	t.Run("refuses to archive a task that is already archived", func(t *testing.T) {
		archivedAt := time.Now().Add(-time.Hour)
		alreadyArchived := newTask()
		alreadyArchived.ArchivedAt = &archivedAt

		result, err := alreadyArchived.Archive()

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTaskAlreadyArchived)
		assert.Equal(t, archivedAt, *alreadyArchived.ArchivedAt, "the original archive date must survive")
	})
}

func TestTaskRestore(t *testing.T) {
	pendingID := uuid.New()
	doingID := uuid.New()
	doneID := uuid.New()

	project := &Project{
		Id: uuid.New(),
		Columns: []ProjectColumn{
			{Id: pendingID, Name: "Pending"},
			{Id: doingID, Name: "Doing"},
			{Id: doneID, Name: "Done", IsDoneColumn: true},
		},
	}

	archivedAt := time.Now().Add(-time.Hour)
	doneAt := time.Now().Add(-2 * time.Hour)

	newArchivedTask := func() *Task {
		return &Task{
			Id:              uuid.New(),
			ProjectId:       project.Id,
			Title:           "Archived task",
			Status:          TaskStatusArchived,
			ProjectColumnId: pendingID,
			Version:         3,
			ArchivedAt:      &archivedAt,
			DoneAt:          &doneAt,
			Updates:         []TaskUpdate{{UpdateType: TaskUpdateTypeArchived}},
		}
	}

	t.Run("puts the task back in the chosen column", func(t *testing.T) {
		task := newArchivedTask()

		restored, err := task.Restore(project, doingID)

		require.NoError(t, err)
		assert.Equal(t, doingID, restored.ProjectColumnId)
		require.NotNil(t, restored.ProjectColumn)
		assert.Equal(t, "Doing", restored.ProjectColumn.Name)
		assert.Equal(t, TaskStatusDoing, restored.Status)
		assert.Nil(t, restored.ArchivedAt)
		assert.Empty(t, restored.Updates)
		assert.Equal(t, 3, restored.Version)
	})

	t.Run("clears the completion when the column does not complete tasks", func(t *testing.T) {
		task := newArchivedTask()

		restored, err := task.Restore(project, doingID)

		require.NoError(t, err)
		assert.Nil(t, restored.DoneAt)
	})

	t.Run("completes the task when restored into the done column", func(t *testing.T) {
		task := newArchivedTask()

		restored, err := task.Restore(project, doneID)

		require.NoError(t, err)
		assert.Equal(t, TaskStatusDone, restored.Status)
		require.NotNil(t, restored.DoneAt)
		assert.WithinDuration(t, time.Now(), *restored.DoneAt, time.Minute)
	})

	t.Run("does not touch the task it was called on", func(t *testing.T) {
		task := newArchivedTask()

		_, err := task.Restore(project, doingID)

		require.NoError(t, err)
		assert.Equal(t, pendingID, task.ProjectColumnId)
		assert.Equal(t, TaskStatusArchived, task.Status)
		require.NotNil(t, task.ArchivedAt)
	})

	t.Run("fails when the task is not archived", func(t *testing.T) {
		task := newArchivedTask()
		task.ArchivedAt = nil

		restored, err := task.Restore(project, doingID)

		assert.Nil(t, restored)
		assert.ErrorIs(t, err, ErrTaskNotArchived)
	})

	t.Run("fails when the column does not belong to the project", func(t *testing.T) {
		task := newArchivedTask()

		restored, err := task.Restore(project, uuid.New())

		assert.Nil(t, restored)
		assert.ErrorIs(t, err, ErrProjectColumnNotFound)
	})

	t.Run("reports the archived state before the column", func(t *testing.T) {
		task := newArchivedTask()
		task.ArchivedAt = nil

		_, err := task.Restore(project, uuid.New())

		assert.ErrorIs(t, err, ErrTaskNotArchived)
	})
}

func TestTaskMoveTo(t *testing.T) {
	doingColumn := &ProjectColumn{Id: uuid.New(), Name: "Doing"}
	doneColumn := &ProjectColumn{Id: uuid.New(), Name: "Done", IsDoneColumn: true}

	task := &Task{
		Id:              uuid.New(),
		ProjectId:       uuid.New(),
		Title:           "Task",
		Description:     "Description",
		Status:          TaskStatusPending,
		ProjectColumnId: uuid.New(),
		Priority:        TaskPriorityHigh,
		Order:           "500000000000",
		Version:         4,
		Tags:            []string{"backend"},
		CreatedAt:       time.Now().Add(-time.Hour),
	}

	t.Run("places the task in the column at the given order", func(t *testing.T) {
		moved := task.MoveTo(doingColumn, "625000000000")

		assert.Equal(t, task.Id, moved.TaskId)
		assert.Equal(t, task.ProjectId, moved.ProjectId)
		assert.Equal(t, "625000000000", moved.Order)
		assert.Equal(t, doingColumn.Id, moved.ProjectColumnId)
		assert.Nil(t, moved.DoneAt)
	})

	t.Run("completes the task when moved into the done column", func(t *testing.T) {
		moved := task.MoveTo(doneColumn, "625000000000")

		require.NotNil(t, moved.DoneAt)
		assert.WithinDuration(t, time.Now(), *moved.DoneAt, time.Minute)
	})

	t.Run("keeps the completion when reordered inside the done column", func(t *testing.T) {
		doneAt := time.Now().Add(-72 * time.Hour)
		doneTask := *task
		doneTask.ProjectColumnId = doneColumn.Id
		doneTask.DoneAt = &doneAt

		moved := doneTask.MoveTo(doneColumn, "625000000000")

		require.NotNil(t, moved.DoneAt)
		assert.Equal(t, doneAt, *moved.DoneAt, "reordering must not rewrite when the task was finished")
	})

	t.Run("clears the completion when moved out of the done column", func(t *testing.T) {
		doneAt := time.Now().Add(-72 * time.Hour)
		doneTask := *task
		doneTask.ProjectColumnId = doneColumn.Id
		doneTask.DoneAt = &doneAt

		moved := doneTask.MoveTo(doingColumn, "625000000000")

		assert.Nil(t, moved.DoneAt)
	})

	t.Run("does not touch the task it was called on", func(t *testing.T) {
		task.MoveTo(doingColumn, "625000000000")

		assert.Equal(t, TaskStatusPending, task.Status)
		assert.Equal(t, "500000000000", task.Order)
	})
}

func TestTaskIsArchived(t *testing.T) {
	archivedAt := time.Now()

	assert.False(t, (&Task{}).IsArchived())
	assert.True(t, (&Task{ArchivedAt: &archivedAt}).IsArchived())
}

func TestTaskApplyEdit(t *testing.T) {
	doingColumn := &ProjectColumn{Id: uuid.New(), Name: "Doing"}
	doneColumn := &ProjectColumn{Id: uuid.New(), Name: "Done", IsDoneColumn: true}

	authorID := uuid.New()
	responsibleID := uuid.New()
	dependencyID := uuid.New()
	createdAt := time.Now().Add(-24 * time.Hour)
	oldDoneAt := time.Now().Add(-time.Hour)
	dueDate := time.Now().Add(48 * time.Hour)

	newTask := func() *Task {
		return &Task{
			Id:              uuid.New(),
			ProjectId:       uuid.New(),
			ProjectColumnId: doneColumn.Id,
			AuthorId:        authorID,
			Author:          &User{Id: authorID, Name: "Author"},
			Title:           "Old title",
			Description:     "Old description",
			Code:            "TASK-1",
			Status:          TaskStatusDone,
			Priority:        TaskPriorityLow,
			Order:           "500000000000",
			Version:         9,
			DoneAt:          &oldDoneAt,
			Tags:            []string{"old"},
			CreatedAt:       createdAt,
			Updates:         []TaskUpdate{{UpdateType: TaskUpdateTypeCreated}},
		}
	}

	dependency := TaskDependencyRef{Id: dependencyID, Title: "Blocker", Code: "TASK-9"}

	fullEdit := TaskEdit{
		Title:          "New title",
		Description:    "New description",
		Priority:       TaskPriorityHigh,
		DueDate:        &dueDate,
		ResponsibleId:  &responsibleID,
		Responsible:    &User{Id: responsibleID, Name: "Responsible"},
		Tags:           []string{"backend", "  ", "", "urgent"},
		DependsOnTasks: []TaskDependencyRef{dependency},
	}

	t.Run("applies the edited fields", func(t *testing.T) {
		task := newTask()

		updated := task.ApplyEdit(fullEdit, doingColumn)

		assert.Equal(t, "New title", updated.Title)
		assert.Equal(t, "New description", updated.Description)
		assert.Equal(t, TaskPriorityHigh, updated.Priority)
		assert.Equal(t, &dueDate, updated.DueDate)
		assert.Equal(t, &responsibleID, updated.ResponsibleId)
		require.NotNil(t, updated.Responsible)
		assert.Equal(t, "Responsible", updated.Responsible.Name)
	})

	t.Run("drops tags that carry no text", func(t *testing.T) {
		task := newTask()

		updated := task.ApplyEdit(fullEdit, doingColumn)

		assert.Equal(t, []string{"backend", "urgent"}, updated.Tags)
	})

	t.Run("derives the dependency ids from the resolved tasks", func(t *testing.T) {
		task := newTask()

		updated := task.ApplyEdit(fullEdit, doingColumn)

		assert.Equal(t, []TaskDependencyRef{dependency}, updated.DependsOnTasks)
		assert.Equal(t, []uuid.UUID{dependencyID}, updated.DependsOnTaskIds)
	})

	t.Run("keeps the fields a task owns for its lifetime", func(t *testing.T) {
		task := newTask()

		updated := task.ApplyEdit(fullEdit, doingColumn)

		assert.Equal(t, task.Id, updated.Id)
		assert.Equal(t, task.ProjectId, updated.ProjectId)
		assert.Equal(t, task.AuthorId, updated.AuthorId)
		assert.Equal(t, task.Author, updated.Author)
		assert.Equal(t, task.Order, updated.Order)
		assert.Equal(t, createdAt, updated.CreatedAt)
		assert.WithinDuration(t, time.Now(), updated.UpdatedAt, time.Minute)
		assert.Empty(t, updated.Updates)
	})

	t.Run("leaves the version to the database", func(t *testing.T) {
		task := newTask()

		updated := task.ApplyEdit(fullEdit, doingColumn)

		assert.Zero(t, updated.Version)
	})

	t.Run("follows the column for status and completion", func(t *testing.T) {
		task := newTask()

		updated := task.ApplyEdit(fullEdit, doingColumn)

		assert.Equal(t, doingColumn.Id, updated.ProjectColumnId)
		assert.Equal(t, doingColumn, updated.ProjectColumn)
		assert.Equal(t, TaskStatusDoing, updated.Status)
		assert.Nil(t, updated.DoneAt, "moving out of the done column uncompletes the task")
	})

	t.Run("completes the task when edited into the done column", func(t *testing.T) {
		task := newTask()
		task.DoneAt = nil

		updated := task.ApplyEdit(fullEdit, doneColumn)

		assert.Equal(t, TaskStatusDone, updated.Status)
		require.NotNil(t, updated.DoneAt)
		assert.WithinDuration(t, time.Now(), *updated.DoneAt, time.Minute)
	})

	t.Run("keeps the completion of a task edited inside the done column", func(t *testing.T) {
		task := newTask()

		updated := task.ApplyEdit(fullEdit, doneColumn)

		require.NotNil(t, updated.DoneAt)
		assert.Equal(t, oldDoneAt, *updated.DoneAt, "editing must not rewrite when the task was finished")
	})

	t.Run("keeps the current code when the edit leaves it out", func(t *testing.T) {
		task := newTask()
		edit := fullEdit
		edit.Code = nil

		updated := task.ApplyEdit(edit, doingColumn)

		assert.Equal(t, "TASK-1", updated.Code)
	})

	t.Run("trims the code the edit provides", func(t *testing.T) {
		task := newTask()
		edit := fullEdit
		code := "  TASK-42  "
		edit.Code = &code

		updated := task.ApplyEdit(edit, doingColumn)

		assert.Equal(t, "TASK-42", updated.Code)
	})

	t.Run("keeps an archived task archived, status included", func(t *testing.T) {
		task := newTask()
		archivedAt := time.Now().Add(-time.Hour)
		task.ArchivedAt = &archivedAt

		updated := task.ApplyEdit(fullEdit, doingColumn)

		assert.Equal(t, &archivedAt, updated.ArchivedAt)
		assert.Equal(t, TaskStatusArchived, updated.Status,
			"an archived task must not report the status of the column it sits in")
	})

	t.Run("does not touch the task it was called on", func(t *testing.T) {
		task := newTask()

		task.ApplyEdit(fullEdit, doingColumn)

		assert.Equal(t, "Old title", task.Title)
		assert.Equal(t, doneColumn.Id, task.ProjectColumnId)
		assert.Equal(t, 9, task.Version)
		assert.Len(t, task.Updates, 1)
	})
}

func TestTaskIsAssignedTo(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name     string
		task     Task
		expected bool
	}{
		{name: "assigned to the user", task: Task{ResponsibleId: &userID}, expected: true},
		{name: "assigned to somebody else", task: Task{ResponsibleId: &otherUserID}, expected: false},
		{name: "not assigned", task: Task{}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.task.IsAssignedTo(userID))
		})
	}
}

func TestTaskIsCompletedIn(t *testing.T) {
	doneColumn := &ProjectColumn{Id: uuid.New(), Name: "Done", IsDoneColumn: true}
	doneAt := time.Now()

	tests := []struct {
		name     string
		task     Task
		expected bool
	}{
		{
			name:     "in the done column and completed",
			task:     Task{ProjectColumnId: doneColumn.Id, DoneAt: &doneAt},
			expected: true,
		},
		{
			name:     "in the done column but never completed",
			task:     Task{ProjectColumnId: doneColumn.Id},
			expected: false,
		},
		{
			name:     "completed before but moved out of the column",
			task:     Task{ProjectColumnId: uuid.New(), DoneAt: &doneAt},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.task.IsCompletedIn(doneColumn))
		})
	}
}

func TestNormalizeTaskTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected []string
	}{
		{name: "no tags at all", tags: nil, expected: []string{}},
		{name: "keeps every meaningful tag", tags: []string{"backend", "urgent"}, expected: []string{"backend", "urgent"}},
		{name: "drops empty and blank tags", tags: []string{"backend", "", "   ", "\t", "urgent"}, expected: []string{"backend", "urgent"}},
		{name: "keeps the surrounding spaces of a tag", tags: []string{" backend "}, expected: []string{" backend "}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NormalizeTaskTags(tt.tags))
		})
	}
}
