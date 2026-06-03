package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskUpdate(t *testing.T) {
	authorID := uuid.New()
	taskID := uuid.New()
	pendingStatusID := uuid.New()
	doingStatusID := uuid.New()
	responsibleID := uuid.New()
	oldResponsibleID := uuid.New()
	newResponsibleID := uuid.New()
	dueDate := time.Now().UTC().Truncate(time.Second)
	oldDueDate := time.Date(2026, time.April, 3, 15, 0, 0, 0, time.UTC)
	sameInstantDifferentTimezone := oldDueDate.In(time.FixedZone("BRT", -3*60*60))

	tests := []struct {
		name          string
		oldTask       *Task
		newTask       *Task
		assertChanges func(t *testing.T, update TaskUpdate)
	}{
		{
			name: "primitive field change",
			oldTask: &Task{
				Id:              taskID,
				Title:           "Old title",
				Description:     "Description",
				ProjectColumnId: pendingStatusID,
				Priority:        TaskPriorityLow,
			},
			newTask: &Task{
				Id:              taskID,
				Title:           "New title",
				Description:     "Description",
				ProjectColumnId: pendingStatusID,
				Priority:        TaskPriorityLow,
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				require.Len(t, update.Changes, 1)
				change := update.Changes[0]
				assert.Equal(t, TaskUpdateTypeUpdated, update.UpdateType)
				assert.Equal(t, "title", change.Field)
				assert.Equal(t, "Old title", change.OldValue)
				assert.Equal(t, "New title", change.NewValue)
				assert.Nil(t, change.OldValueId)
				assert.Nil(t, change.NewValueId)
				assert.Nil(t, change.OldDisplayValue)
				assert.Nil(t, change.NewDisplayValue)
			},
		},
		{
			name: "code change",
			oldTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
				Code:            "TASK-1",
			},
			newTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
				Code:            "TASK-2",
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				require.Len(t, update.Changes, 1)
				change := update.Changes[0]
				assert.Equal(t, TaskUpdateTypeUpdated, update.UpdateType)
				assert.Equal(t, "code", change.Field)
				assert.Equal(t, "TASK-1", change.OldValue)
				assert.Equal(t, "TASK-2", change.NewValue)
			},
		},
		{
			name: "assigned responsible",
			oldTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
			},
			newTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
				ResponsibleId:   &responsibleID,
				Responsible:     &User{Id: responsibleID, Name: "Maria"},
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				require.Len(t, update.Changes, 1)
				change := update.Changes[0]
				require.NotNil(t, change.NewValueId)
				require.NotNil(t, change.NewDisplayValue)
				require.NotNil(t, change.SubjectId)
				assert.Equal(t, TaskUpdateTypeAssigned, update.UpdateType)
				assert.Equal(t, "responsible_id", change.Field)
				assert.Equal(t, responsibleID.String(), change.NewValue)
				assert.Equal(t, responsibleID, *change.NewValueId)
				assert.Equal(t, "Maria", *change.NewDisplayValue)
				assert.Equal(t, responsibleID, *change.SubjectId)
				assert.Equal(t, "", change.OldValue)
				assert.Nil(t, change.OldValueId)
				assert.Nil(t, change.OldDisplayValue)
			},
		},
		{
			name: "unassigned responsible",
			oldTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
				ResponsibleId:   &responsibleID,
				Responsible:     &User{Id: responsibleID, Name: "Maria"},
			},
			newTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				require.Len(t, update.Changes, 1)
				change := update.Changes[0]
				require.NotNil(t, change.OldValueId)
				require.NotNil(t, change.OldDisplayValue)
				require.NotNil(t, change.SubjectId)
				assert.Equal(t, TaskUpdateTypeUnassigned, update.UpdateType)
				assert.Equal(t, responsibleID.String(), change.OldValue)
				assert.Equal(t, responsibleID, *change.OldValueId)
				assert.Equal(t, "Maria", *change.OldDisplayValue)
				assert.Equal(t, responsibleID, *change.SubjectId)
				assert.Equal(t, "", change.NewValue)
				assert.Nil(t, change.NewValueId)
				assert.Nil(t, change.NewDisplayValue)
			},
		},
		{
			name: "reassigned responsible",
			oldTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
				ResponsibleId:   &oldResponsibleID,
				Responsible:     &User{Id: oldResponsibleID, Name: "Maria"},
			},
			newTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
				ResponsibleId:   &newResponsibleID,
				Responsible:     &User{Id: newResponsibleID, Name: "Joao"},
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				require.Len(t, update.Changes, 1)
				change := update.Changes[0]
				require.NotNil(t, change.OldValueId)
				require.NotNil(t, change.NewValueId)
				require.NotNil(t, change.OldDisplayValue)
				require.NotNil(t, change.NewDisplayValue)
				assert.Equal(t, TaskUpdateTypeUpdated, update.UpdateType)
				assert.Equal(t, oldResponsibleID, *change.OldValueId)
				assert.Equal(t, newResponsibleID, *change.NewValueId)
				assert.Equal(t, "Maria", *change.OldDisplayValue)
				assert.Equal(t, "Joao", *change.NewDisplayValue)
				assert.Nil(t, change.SubjectId)
			},
		},
		{
			name: "mixed field edit with responsible",
			oldTask: &Task{
				Id:              taskID,
				Title:           "Old title",
				Description:     "Description",
				ProjectColumnId: pendingStatusID,
			},
			newTask: &Task{
				Id:              taskID,
				Title:           "New title",
				Description:     "Description",
				ProjectColumnId: pendingStatusID,
				ResponsibleId:   &responsibleID,
				Responsible:     &User{Id: responsibleID, Name: "Maria"},
				DueDate:         &dueDate,
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				require.Len(t, update.Changes, 3)
				assert.Equal(t, TaskUpdateTypeUpdated, update.UpdateType)

				fields := []string{update.Changes[0].Field, update.Changes[1].Field, update.Changes[2].Field}
				assert.Equal(t, []string{"title", "responsible_id", "due_date"}, fields)
			},
		},
		{
			name: "no changes",
			oldTask: &Task{
				Id:              taskID,
				Title:           "Same title",
				Description:     "Same description",
				ProjectColumnId: pendingStatusID,
				Priority:        TaskPriorityLow,
			},
			newTask: &Task{
				Id:              taskID,
				Title:           "Same title",
				Description:     "Same description",
				ProjectColumnId: pendingStatusID,
				Priority:        TaskPriorityLow,
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				assert.Empty(t, update.Changes)
				assert.Equal(t, TaskUpdateTypeUpdated, update.UpdateType)
			},
		},
		{
			name: "due date same instant different timezone",
			oldTask: &Task{
				Id:      taskID,
				DueDate: &oldDueDate,
			},
			newTask: &Task{
				Id:      taskID,
				DueDate: &sameInstantDifferentTimezone,
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				assert.Empty(t, update.Changes)
				assert.Equal(t, TaskUpdateTypeUpdated, update.UpdateType)
			},
		},
		{
			name: "status change uses display names",
			oldTask: &Task{
				Id:              taskID,
				ProjectColumnId: pendingStatusID,
				ProjectColumn:   &ProjectColumn{Id: pendingStatusID, Name: "Pending", Color: "#64748B"},
			},
			newTask: &Task{
				Id:              taskID,
				ProjectColumnId: doingStatusID,
				ProjectColumn:   &ProjectColumn{Id: doingStatusID, Name: "Doing", Color: "#2563EB"},
			},
			assertChanges: func(t *testing.T, update TaskUpdate) {
				require.Len(t, update.Changes, 1)
				change := update.Changes[0]
				require.NotNil(t, change.OldDisplayValue)
				require.NotNil(t, change.NewDisplayValue)
				assert.Equal(t, TaskUpdateTypeColumn, update.UpdateType)
				assert.Equal(t, pendingStatusID.String(), change.OldValue)
				assert.Equal(t, doingStatusID.String(), change.NewValue)
				assert.Equal(t, "Pending", *change.OldDisplayValue)
				assert.Equal(t, "Doing", *change.NewDisplayValue)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := NewTaskUpdate(tt.oldTask, tt.newTask, &User{Id: authorID})
			tt.assertChanges(t, update)
		})
	}
}
