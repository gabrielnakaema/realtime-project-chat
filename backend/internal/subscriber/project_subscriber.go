package subscriber

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
)

type ProjectNotifier interface {
	SendUpdatedProject(context.Context, *domain.Project) error
}

type ProjectSubscriber struct {
	logger     *slog.Logger
	subscriber *Subscriber
	notifier   ProjectNotifier
}

func NewProjectSubscriber(ctx context.Context, config *config.Config, logger *slog.Logger, notifier ProjectNotifier) (*ProjectSubscriber, error) {
	subscriber, err := NewSubscriber(config, "project.subscriber")
	if err != nil {
		return nil, err
	}

	projectSubscriber := &ProjectSubscriber{
		logger:     logger,
		subscriber: subscriber,
		notifier:   notifier,
	}

	topics := []events.Topic{events.ProjectUpdated}

	err = subscriber.Subscribe(ctx, topics, projectSubscriber.handleProjectEvents, projectSubscriber.logger)
	if err != nil {
		return nil, err
	}

	return projectSubscriber, nil
}

func (ps *ProjectSubscriber) Close() error {
	return ps.subscriber.Close()
}

func (ps *ProjectSubscriber) handleProjectEvents(ctx context.Context, message Message) error {
	switch message.Topic {
	case events.ProjectUpdated:
		return ps.handleProjectUpdated(ctx, message)
	default:
		return nil
	}
}

func (ps *ProjectSubscriber) handleProjectUpdated(ctx context.Context, message Message) error {
	var payload events.ProjectUpdatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal project updated payload", err)
	}

	err = ps.notifier.SendUpdatedProject(ctx, &payload.Project)
	if err != nil {
		return domain.ServerError("failed to send updated project to ws server", err)
	}

	return nil
}
