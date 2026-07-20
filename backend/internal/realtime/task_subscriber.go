package realtime

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/config"
	"github.com/gabrielnakaema/project-chat/internal/platform/messaging"
	"github.com/google/uuid"
)

type TaskNotifier interface {
	SendCreatedTask(context.Context, *domain.Task) error
	SendUpdatedTask(context.Context, *domain.Task, *uuid.UUID) error
	SendCreatedTaskComment(context.Context, *domain.TaskComment) error
}

type TaskSubscriber struct {
	logger     *slog.Logger
	subscriber *messaging.Subscriber
	notifier   TaskNotifier
}

func NewTaskSubscriber(ctx context.Context, config *config.Config, logger *slog.Logger, notifier TaskNotifier) (*TaskSubscriber, error) {
	subscriber, err := messaging.NewSubscriber(config, "task.subscriber")
	if err != nil {
		return nil, err
	}

	taskSubscriber := &TaskSubscriber{
		logger:     logger,
		subscriber: subscriber,
		notifier:   notifier,
	}

	topics := []events.Topic{events.TaskCreated, events.TaskUpdated, events.TaskCommentCreated}

	err = subscriber.Subscribe(ctx, topics, taskSubscriber.handleTaskEvents, taskSubscriber.logger)
	if err != nil {
		return nil, err
	}

	return taskSubscriber, nil
}

func (ts *TaskSubscriber) Close() error {
	return ts.subscriber.Close()
}

func (ts *TaskSubscriber) handleTaskEvents(ctx context.Context, message messaging.Message) error {
	switch message.Topic {
	case events.TaskCreated:
		return ts.handleTaskCreated(ctx, message)
	case events.TaskUpdated:
		return ts.handleTaskUpdated(ctx, message)
	case events.TaskCommentCreated:
		return ts.handleTaskCommentCreated(ctx, message)
	default:
		return nil

	}
}

func (ts *TaskSubscriber) handleTaskCommentCreated(ctx context.Context, message messaging.Message) error {
	var payload events.TaskCommentCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return apperr.ServerError("failed to unmarshal task comment", err)
	}

	err = ts.notifier.SendCreatedTaskComment(ctx, &payload.TaskComment)
	if err != nil {
		return apperr.ServerError("failed to send created task comment to ws server", err)
	}

	return nil
}

func (ts *TaskSubscriber) handleTaskCreated(ctx context.Context, message messaging.Message) error {
	var payload events.TaskCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return apperr.ServerError("failed to unmarshal task", err)
	}

	err = ts.notifier.SendCreatedTask(ctx, &payload.Task)
	if err != nil {
		return apperr.ServerError("failed to send created task to ws server", err)
	}

	return nil
}

func (ts *TaskSubscriber) handleTaskUpdated(ctx context.Context, message messaging.Message) error {
	var payload events.TaskUpdatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return apperr.ServerError("failed to unmarshal task", err)
	}

	err = ts.notifier.SendUpdatedTask(ctx, &payload.Task, payload.PreviousProjectColumnID)
	if err != nil {
		return apperr.ServerError("failed to send updated task to ws server", err)
	}

	return nil
}
