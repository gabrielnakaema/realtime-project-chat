package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDedupeTaskIDs(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	third := uuid.New()

	tests := []struct {
		name     string
		ids      []uuid.UUID
		expected []uuid.UUID
	}{
		{name: "no ids at all", ids: nil, expected: []uuid.UUID{}},
		{name: "empty list", ids: []uuid.UUID{}, expected: []uuid.UUID{}},
		{name: "already unique, order kept", ids: []uuid.UUID{third, first, second}, expected: []uuid.UUID{third, first, second}},
		{name: "keeps the first occurrence", ids: []uuid.UUID{first, second, first, third, second}, expected: []uuid.UUID{first, second, third}},
		{name: "all the same", ids: []uuid.UUID{first, first, first}, expected: []uuid.UUID{first}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, DedupeTaskIDs(tt.ids))
		})
	}
}

func TestValidateDependencyIDs(t *testing.T) {
	taskID := uuid.New()
	otherID := uuid.New()

	tests := []struct {
		name        string
		taskID      uuid.UUID
		dependsOn   []uuid.UUID
		expectedErr error
	}{
		{name: "no dependencies", taskID: taskID, dependsOn: nil},
		{name: "dependencies on other tasks", taskID: taskID, dependsOn: []uuid.UUID{otherID, uuid.New()}},
		{name: "nil id", taskID: taskID, dependsOn: []uuid.UUID{otherID, uuid.Nil}, expectedErr: ErrInvalidDependencyTaskId},
		{name: "task depending on itself", taskID: taskID, dependsOn: []uuid.UUID{otherID, taskID}, expectedErr: ErrSelfDependency},
		{name: "task being created cannot be depended on yet", taskID: uuid.Nil, dependsOn: []uuid.UUID{otherID}},
		{name: "task being created still rejects a nil dependency", taskID: uuid.Nil, dependsOn: []uuid.UUID{uuid.Nil}, expectedErr: ErrInvalidDependencyTaskId},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependencyIDs(tt.taskID, tt.dependsOn)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestResolveDependencyRefs(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()

	firstRef := TaskDependencyRef{Id: firstID, Title: "First", Code: "TASK-1"}
	secondRef := TaskDependencyRef{Id: secondID, Title: "Second", Code: "TASK-2"}

	t.Run("returns the refs in the requested order", func(t *testing.T) {
		refs, err := ResolveDependencyRefs([]uuid.UUID{secondID, firstID}, []TaskDependencyRef{firstRef, secondRef})

		require.NoError(t, err)
		assert.Equal(t, []TaskDependencyRef{secondRef, firstRef}, refs)
	})

	t.Run("returns nothing when nothing was requested", func(t *testing.T) {
		refs, err := ResolveDependencyRefs(nil, []TaskDependencyRef{firstRef})

		require.NoError(t, err)
		assert.Empty(t, refs)
	})

	t.Run("fails when an id has no matching task", func(t *testing.T) {
		refs, err := ResolveDependencyRefs([]uuid.UUID{firstID, uuid.New()}, []TaskDependencyRef{firstRef})

		assert.Nil(t, refs)
		assert.ErrorIs(t, err, ErrUnknownDependencyTasks)
	})
}

func TestHasDependencyCycle(t *testing.T) {
	taskID := uuid.New()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()

	tests := []struct {
		name      string
		dependsOn []uuid.UUID
		edges     []TaskDependencyEdge
		expected  bool
	}{
		{
			name:      "no dependencies",
			dependsOn: nil,
			edges:     []TaskDependencyEdge{{TaskId: a, DependsOnTaskId: taskID}},
			expected:  false,
		},
		{
			name:      "independent dependencies",
			dependsOn: []uuid.UUID{a, b},
			edges:     []TaskDependencyEdge{{TaskId: b, DependsOnTaskId: c}},
			expected:  false,
		},
		{
			name:      "direct cycle: the dependency already depends on the task",
			dependsOn: []uuid.UUID{a},
			edges:     []TaskDependencyEdge{{TaskId: a, DependsOnTaskId: taskID}},
			expected:  true,
		},
		{
			name:      "indirect cycle through two hops",
			dependsOn: []uuid.UUID{a},
			edges: []TaskDependencyEdge{
				{TaskId: a, DependsOnTaskId: b},
				{TaskId: b, DependsOnTaskId: c},
				{TaskId: c, DependsOnTaskId: taskID},
			},
			expected: true,
		},
		{
			name:      "cycle reached through the second dependency only",
			dependsOn: []uuid.UUID{a, b},
			edges:     []TaskDependencyEdge{{TaskId: b, DependsOnTaskId: taskID}},
			expected:  true,
		},
		{
			name:      "the task's own existing edges are replaced, not added",
			dependsOn: []uuid.UUID{c},
			edges: []TaskDependencyEdge{
				{TaskId: taskID, DependsOnTaskId: a},
				{TaskId: a, DependsOnTaskId: b},
			},
			expected: false,
		},
		{
			name:      "a pre-existing cycle between other tasks is not the task's fault",
			dependsOn: []uuid.UUID{c},
			edges: []TaskDependencyEdge{
				{TaskId: a, DependsOnTaskId: b},
				{TaskId: b, DependsOnTaskId: a},
			},
			expected: false,
		},
		{
			name:      "walks into a pre-existing cycle without hanging",
			dependsOn: []uuid.UUID{a},
			edges: []TaskDependencyEdge{
				{TaskId: a, DependsOnTaskId: b},
				{TaskId: b, DependsOnTaskId: a},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasDependencyCycle(taskID, tt.dependsOn, tt.edges))
		})
	}
}
