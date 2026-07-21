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
)

type ProjectNotifier interface {
	SendUpdatedProject(context.Context, *domain.Project) error
	SendProjectDeleted(context.Context, *domain.Project) error
	SendProjectMemberCreated(context.Context, *domain.ProjectMember) error
	SendProjectMemberRemoved(context.Context, *domain.ProjectMember) error
}

type ProjectSubscriber struct {
	logger     *slog.Logger
	subscriber *messaging.Subscriber
	notifier   ProjectNotifier
}

func NewProjectSubscriber(ctx context.Context, config *config.Config, logger *slog.Logger, notifier ProjectNotifier) (*ProjectSubscriber, error) {
	subscriber, err := messaging.NewSubscriber(config, "project.subscriber")
	if err != nil {
		return nil, err
	}

	projectSubscriber := &ProjectSubscriber{
		logger:     logger,
		subscriber: subscriber,
		notifier:   notifier,
	}

	topics := []events.Topic{events.ProjectUpdated, events.ProjectDeleted, events.ProjectMemberCreated, events.ProjectMemberRemoved}

	err = subscriber.Subscribe(ctx, topics, projectSubscriber.handleProjectEvents, projectSubscriber.logger)
	if err != nil {
		return nil, err
	}

	return projectSubscriber, nil
}

func (ps *ProjectSubscriber) Close() error {
	return ps.subscriber.Close()
}

func (ps *ProjectSubscriber) handleProjectEvents(ctx context.Context, message messaging.Message) error {
	switch message.Topic {
	case events.ProjectUpdated:
		return ps.handleProjectUpdated(ctx, message)
	case events.ProjectDeleted:
		return ps.handleProjectDeleted(ctx, message)
	case events.ProjectMemberCreated:
		return ps.handleProjectMemberCreated(ctx, message)
	case events.ProjectMemberRemoved:
		return ps.handleProjectMemberRemoved(ctx, message)
	default:
		return nil
	}
}

func (ps *ProjectSubscriber) handleProjectMemberCreated(ctx context.Context, message messaging.Message) error {
	var payload events.ProjectMemberCreatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return apperr.ServerError("failed to unmarshal project member created payload", err)
	}

	if err := ps.notifier.SendProjectMemberCreated(ctx, &payload.ProjectMember); err != nil {
		return apperr.ServerError("failed to send created project member to ws server", err)
	}

	return nil
}

func (ps *ProjectSubscriber) handleProjectMemberRemoved(ctx context.Context, message messaging.Message) error {
	var payload events.ProjectMemberRemovedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return apperr.ServerError("failed to unmarshal project member removed payload", err)
	}

	if err := ps.notifier.SendProjectMemberRemoved(ctx, &payload.ProjectMember); err != nil {
		return apperr.ServerError("failed to send removed project member to ws server", err)
	}

	return nil
}

func (ps *ProjectSubscriber) handleProjectDeleted(ctx context.Context, message messaging.Message) error {
	var payload events.ProjectDeletedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return apperr.ServerError("failed to unmarshal project deleted payload", err)
	}

	if err := ps.notifier.SendProjectDeleted(ctx, &payload.Project); err != nil {
		return apperr.ServerError("failed to send deleted project to ws server", err)
	}

	return nil
}

func (ps *ProjectSubscriber) handleProjectUpdated(ctx context.Context, message messaging.Message) error {
	var payload events.ProjectUpdatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return apperr.ServerError("failed to unmarshal project updated payload", err)
	}

	err = ps.notifier.SendUpdatedProject(ctx, &payload.Project)
	if err != nil {
		return apperr.ServerError("failed to send updated project to ws server", err)
	}

	return nil
}
