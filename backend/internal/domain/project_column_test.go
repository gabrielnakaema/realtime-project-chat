package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectColumn(t *testing.T) {
	pendingID := uuid.New()
	doingID := uuid.New()
	doneID := uuid.New()

	project := Project{
		Id: uuid.New(),
		Columns: []ProjectColumn{
			{Id: pendingID, Name: "Pending", Position: 0},
			{Id: doingID, Name: "In Progress", Position: 1},
			{Id: doneID, Name: "Done", Position: 2, IsDoneColumn: true},
		},
	}

	t.Run("returns the column with the given id", func(t *testing.T) {
		column, err := project.Column(doingID)

		require.NoError(t, err)
		assert.Equal(t, doingID, column.Id)
		assert.Equal(t, "In Progress", column.Name)
	})

	t.Run("returns a copy so callers cannot edit the project", func(t *testing.T) {
		column, err := project.Column(pendingID)
		require.NoError(t, err)

		column.Name = "Renamed"

		assert.Equal(t, "Pending", project.Columns[0].Name)
	})

	t.Run("fails for a column of another project", func(t *testing.T) {
		column, err := project.Column(uuid.New())

		assert.Nil(t, column)
		assert.ErrorIs(t, err, ErrProjectColumnNotFound)
	})
}

func TestProjectValidateColumnIDs(t *testing.T) {
	pendingID := uuid.New()
	doneID := uuid.New()

	project := Project{
		Columns: []ProjectColumn{
			{Id: pendingID, Name: "Pending"},
			{Id: doneID, Name: "Done", IsDoneColumn: true},
		},
	}

	tests := []struct {
		name        string
		ids         []uuid.UUID
		expectedErr error
	}{
		{name: "no ids at all", ids: nil},
		{name: "empty list", ids: []uuid.UUID{}},
		{name: "every id belongs to the project", ids: []uuid.UUID{pendingID, doneID}},
		{name: "repeated valid id", ids: []uuid.UUID{pendingID, pendingID}},
		{name: "one foreign id", ids: []uuid.UUID{pendingID, uuid.New()}, expectedErr: ErrProjectColumnNotFound},
		{name: "nil id", ids: []uuid.UUID{uuid.Nil}, expectedErr: ErrProjectColumnNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := project.ValidateColumnIDs(tt.ids)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestProjectDoneColumn(t *testing.T) {
	doneID := uuid.New()

	t.Run("returns the single done column", func(t *testing.T) {
		project := Project{
			Columns: []ProjectColumn{
				{Id: uuid.New(), Name: "Pending"},
				{Id: doneID, Name: "Done", IsDoneColumn: true},
			},
		}

		column, err := project.DoneColumn()

		require.NoError(t, err)
		assert.Equal(t, doneID, column.Id)
	})

	t.Run("fails when no column completes tasks", func(t *testing.T) {
		project := Project{
			Columns: []ProjectColumn{
				{Id: uuid.New(), Name: "Pending"},
			},
		}

		column, err := project.DoneColumn()

		assert.Nil(t, column)
		assert.ErrorIs(t, err, ErrDoneColumnNotUnique)
	})

	t.Run("fails when the done column is ambiguous", func(t *testing.T) {
		project := Project{
			Columns: []ProjectColumn{
				{Id: doneID, Name: "Done", IsDoneColumn: true},
				{Id: uuid.New(), Name: "Verified", IsDoneColumn: true},
			},
		}

		column, err := project.DoneColumn()

		assert.Nil(t, column)
		assert.ErrorIs(t, err, ErrDoneColumnNotUnique)
	})

	t.Run("fails for a project without columns", func(t *testing.T) {
		project := Project{}

		column, err := project.DoneColumn()

		assert.Nil(t, column)
		assert.ErrorIs(t, err, ErrDoneColumnNotUnique)
	})
}

func TestTaskStatusIn(t *testing.T) {
	archivedAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name       string
		column     *ProjectColumn
		archivedAt *time.Time
		expected   TaskStatus
	}{
		{name: "pending", column: &ProjectColumn{Name: "Pending"}, expected: TaskStatusPending},
		{name: "doing", column: &ProjectColumn{Name: "Doing"}, expected: TaskStatusDoing},
		{name: "done", column: &ProjectColumn{Name: "DONE", IsDoneColumn: true}, expected: TaskStatusDone},
		{name: "custom column name", column: &ProjectColumn{Name: "In Review"}, expected: TaskStatus("in review")},
		{
			name:       "archiving wins over the column",
			column:     &ProjectColumn{Name: "Doing"},
			archivedAt: &archivedAt,
			expected:   TaskStatusArchived,
		},
		{
			name:       "archived without a loaded column",
			archivedAt: &archivedAt,
			expected:   TaskStatusArchived,
		},
		{name: "no column to derive from", expected: TaskStatus("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, TaskStatusIn(tt.column, tt.archivedAt))
		})
	}
}

func TestProjectColumnCompletedAt(t *testing.T) {
	earlier := time.Now().Add(-24 * time.Hour)

	t.Run("a done column completes a task that was not completed yet", func(t *testing.T) {
		column := ProjectColumn{Name: "Done", IsDoneColumn: true}

		completedAt := column.CompletedAt(nil)

		require.NotNil(t, completedAt)
		assert.WithinDuration(t, time.Now(), *completedAt, time.Minute)
	})

	t.Run("a done column keeps the completion a task already had", func(t *testing.T) {
		column := ProjectColumn{Name: "Done", IsDoneColumn: true}

		completedAt := column.CompletedAt(&earlier)

		require.NotNil(t, completedAt)
		assert.Equal(t, earlier, *completedAt, "re-completing must not rewrite when the task was finished")
	})

	t.Run("any other column leaves the task uncompleted", func(t *testing.T) {
		column := ProjectColumn{Name: "Doing"}

		assert.Nil(t, column.CompletedAt(nil))
	})

	t.Run("leaving a done column clears an existing completion", func(t *testing.T) {
		column := ProjectColumn{Name: "Doing"}

		assert.Nil(t, column.CompletedAt(&earlier))
	})
}
