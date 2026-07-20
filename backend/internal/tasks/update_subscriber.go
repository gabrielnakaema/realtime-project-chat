package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/subscriber"
)

type TaskUpdateRepository interface {
	CreateUpdates(ctx context.Context, task *domain.Task, updates []domain.TaskUpdate) error
}

type TaskUpdateSubscriber struct {
	logger     *slog.Logger
	subscriber *subscriber.Subscriber
	repository TaskUpdateRepository
}

func NewTaskUpdateSubscriber(ctx context.Context, config *config.Config, logger *slog.Logger, repository TaskUpdateRepository) (*TaskUpdateSubscriber, error) {
	sub, err := subscriber.NewSubscriber(config, "task_update.subscriber")
	if err != nil {
		return nil, err
	}

	taskUpdateSubscriber := &TaskUpdateSubscriber{
		logger:     logger,
		subscriber: sub,
		repository: repository,
	}

	topics := []events.Topic{events.TaskCreated, events.TaskUpdated}
	err = sub.Subscribe(ctx, topics, taskUpdateSubscriber.handleTaskUpdateEvents, taskUpdateSubscriber.logger)
	if err != nil {
		return nil, err
	}

	return taskUpdateSubscriber, nil
}

func (s *TaskUpdateSubscriber) Close() error {
	return s.subscriber.Close()
}

func (s *TaskUpdateSubscriber) handleTaskUpdateEvents(ctx context.Context, message subscriber.Message) error {
	switch message.Topic {
	case events.TaskCreated:
		if err := s.handleTaskCreated(ctx, message); err != nil {
			s.logger.Error("skipping task created update event", "topic", message.Topic, "error", err.Error())
		}
		return nil
	case events.TaskUpdated:
		if err := s.handleTaskUpdated(ctx, message); err != nil {
			s.logger.Error("skipping task updated event", "topic", message.Topic, "error", err.Error())
		}
		return nil
	default:
		return nil
	}
}

func (s *TaskUpdateSubscriber) handleTaskCreated(ctx context.Context, message subscriber.Message) error {
	var payload events.TaskCreatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ServerError("failed to unmarshal task created payload", err)
	}

	update := domain.NewTaskCreatedUpdate(&payload.Task, &payload.User)
	update.ActionOrigin = payload.ActionOrigin.OrUser()

	ctx = domain.WithActionOrigin(ctx, payload.ActionOrigin)
	if err := s.repository.CreateUpdates(ctx, &payload.Task, []domain.TaskUpdate{update}); err != nil {
		return domain.ServerError("failed to create task update for created task", err)
	}

	return nil
}

func (s *TaskUpdateSubscriber) handleTaskUpdated(ctx context.Context, message subscriber.Message) error {
	var payload events.TaskUpdatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ServerError("failed to unmarshal task updated payload", err)
	}

	if payload.PreviousTask == nil {
		s.logger.Warn("task updated payload missing previous task", "task_id", payload.Task.Id, "topic", message.Topic)
		return domain.ServerError("failed to create task update", errors.New("task updated payload missing previous_task"))
	}

	update := domain.NewTaskUpdate(payload.PreviousTask, &payload.Task, &payload.User)
	update.ActionOrigin = payload.ActionOrigin.OrUser()
	if len(update.Changes) == 0 {
		return nil
	}

	ctx = domain.WithActionOrigin(ctx, payload.ActionOrigin)
	if err := s.repository.CreateUpdates(ctx, &payload.Task, []domain.TaskUpdate{update}); err != nil {
		return domain.ServerError("failed to create task update", err)
	}

	return nil
}
