package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type stubTaskUpdateRepository struct {
	err        error
	callCount  int
	lastTask   *domain.Task
	lastUpdate []domain.TaskUpdate
}

func (s *stubTaskUpdateRepository) CreateUpdates(ctx context.Context, task *domain.Task, updates []domain.TaskUpdate) error {
	s.callCount++
	s.lastTask = task
	s.lastUpdate = updates
	return s.err
}

func testTaskUpdateLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestTaskUpdateSubscriberHandleTaskUpdateEvents_SkipsPoisonMessages(t *testing.T) {
	taskID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	oldColumnID := uuid.New()
	newColumnID := uuid.New()

	validTask := domain.Task{
		Id:              taskID,
		ProjectId:       projectID,
		ProjectColumnId: newColumnID,
		ProjectColumn:   &domain.ProjectColumn{Id: newColumnID, Name: "Doing"},
	}
	previousTask := &domain.Task{
		Id:              taskID,
		ProjectId:       projectID,
		ProjectColumnId: oldColumnID,
		ProjectColumn:   &domain.ProjectColumn{Id: oldColumnID, Name: "Pending"},
	}

	t.Run("returns nil when previous_task is missing", func(t *testing.T) {
		repo := &stubTaskUpdateRepository{}
		sub := &TaskUpdateSubscriber{
			logger:     testTaskUpdateLogger(),
			repository: repo,
		}

		payload := &events.TaskUpdatedPayload{
			Task: validTask,
			User: domain.User{Id: userID},
		}

		bytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		err = sub.handleTaskUpdateEvents(context.Background(), Message{
			Topic: events.TaskUpdated,
			Value: bytes,
		})

		assert.NoError(t, err)
		assert.Equal(t, 0, repo.callCount)
	})

	t.Run("returns nil when repository insert fails", func(t *testing.T) {
		repo := &stubTaskUpdateRepository{err: errors.New("insert failed")}
		sub := &TaskUpdateSubscriber{
			logger:     testTaskUpdateLogger(),
			repository: repo,
		}

		payload := &events.TaskUpdatedPayload{
			Task:         validTask,
			PreviousTask: previousTask,
			User:         domain.User{Id: userID},
		}

		bytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		err = sub.handleTaskUpdateEvents(context.Background(), Message{
			Topic: events.TaskUpdated,
			Value: bytes,
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, repo.callCount)
	})
}
