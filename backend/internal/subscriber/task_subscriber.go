package subscriber

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
)

type TaskNotifier interface {
	SendCreatedTask(context.Context, *domain.Task) error
	SendUpdatedTask(context.Context, *domain.Task) error
}

type TaskSubscriber struct {
	logger     *slog.Logger
	subscriber *Subscriber
	notifier   TaskNotifier
}

func NewTaskSubscriber(config *config.Config, logger *slog.Logger, notifier TaskNotifier) (*TaskSubscriber, error) {
	subscriber, err := NewSubscriber(config, "task.subscriber")
	if err != nil {
		return nil, err
	}

	taskSubscriber := &TaskSubscriber{
		logger:     logger,
		subscriber: subscriber,
		notifier:   notifier,
	}

	topics := []events.Topic{events.TaskCreated, events.TaskUpdated}

	err = subscriber.Subscribe(context.Background(), topics, taskSubscriber.handleTaskEvents, taskSubscriber.logger)
	if err != nil {
		return nil, err
	}

	return taskSubscriber, nil

}

func (ts *TaskSubscriber) handleTaskEvents(ctx context.Context, message Message) error {
	switch message.Topic {
	case events.TaskCreated:
		return ts.handleTaskCreated(ctx, message)
	case events.TaskUpdated:
		return ts.handleTaskUpdated(ctx, message)
	default:
		return nil

	}
}

func (ts *TaskSubscriber) handleTaskCreated(ctx context.Context, message Message) error {
	var task domain.Task
	err := json.Unmarshal(message.Value, &task)
	if err != nil {
		return domain.ServerError("failed to unmarshal task", err)
	}

	err = ts.notifier.SendCreatedTask(ctx, &task)
	if err != nil {
		return domain.ServerError("failed to send created task to ws server", err)
	}

	return nil
}

func (ts *TaskSubscriber) handleTaskUpdated(ctx context.Context, message Message) error {
	var task domain.Task
	err := json.Unmarshal(message.Value, &task)
	if err != nil {
		return domain.ServerError("failed to unmarshal task", err)
	}

	err = ts.notifier.SendUpdatedTask(ctx, &task)
	if err != nil {
		return domain.ServerError("failed to send updated task to ws server", err)
	}

	return nil
}
