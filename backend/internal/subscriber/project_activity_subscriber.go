package subscriber

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/repository"
	"github.com/google/uuid"
)

type ProjectRepository interface {
	MarkUpdatedAt(ctx context.Context, projectId uuid.UUID) error
}

type ProjectActivitySubscriber struct {
	logger            *slog.Logger
	subscriber        *Subscriber
	repository        *repository.ProjectActivityRepository
	projectRepository ProjectRepository
}

func NewProjectActivitySubscriber(ctx context.Context, config *config.Config, logger *slog.Logger, repository *repository.ProjectActivityRepository, projectRepository ProjectRepository) (*ProjectActivitySubscriber, error) {
	subscriber, err := NewSubscriber(config, "project_activity.subscriber")
	if err != nil {
		return nil, err
	}

	projectActivitySubscriber := &ProjectActivitySubscriber{
		logger:            logger,
		subscriber:        subscriber,
		repository:        repository,
		projectRepository: projectRepository,
	}

	topics := []events.Topic{events.ProjectCreated, events.ProjectUpdated, events.ProjectMemberCreated, events.ProjectMemberRemoved, events.TaskCreated, events.TaskUpdated}
	err = subscriber.Subscribe(ctx, topics, projectActivitySubscriber.handleProjectActivityEvents, projectActivitySubscriber.logger)
	if err != nil {
		return nil, err
	}

	return projectActivitySubscriber, nil
}

func (pas *ProjectActivitySubscriber) Close() error {
	return pas.subscriber.Close()
}

func (pas *ProjectActivitySubscriber) handleProjectActivityEvents(ctx context.Context, message Message) error {
	switch message.Topic {
	case events.ProjectCreated:
		return pas.handleProjectCreated(ctx, message)
	case events.ProjectUpdated:
		return pas.handleProjectUpdated(ctx, message)
	case events.ProjectMemberCreated:
		return pas.handleProjectMemberCreated(ctx, message)
	case events.TaskCreated:
		return pas.handleTaskCreated(ctx, message)
	case events.TaskUpdated:
		return pas.handleTaskUpdated(ctx, message)
	default:
		return nil
	}
}

func (pas *ProjectActivitySubscriber) handleProjectCreated(ctx context.Context, message Message) error {
	var payload events.ProjectCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal project created payload", err)
	}

	activity := domain.ProjectCreatedActivity(payload.Project, payload.User)
	err = pas.repository.Create(ctx, &activity)
	if err != nil {
		return domain.ServerError("failed to create project created activity", err)
	}

	err = pas.projectRepository.MarkUpdatedAt(ctx, payload.Project.Id)
	if err != nil {
		return domain.ServerError("failed to mark project updated at", err)
	}

	return nil
}

func (pas *ProjectActivitySubscriber) handleProjectUpdated(ctx context.Context, message Message) error {
	var payload events.ProjectUpdatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal project updated payload", err)
	}

	activity := domain.ProjectUpdatedActivity(payload.Project, payload.User)
	err = pas.repository.Create(ctx, &activity)
	if err != nil {
		return domain.ServerError("failed to create project updated activity", err)
	}

	err = pas.projectRepository.MarkUpdatedAt(ctx, payload.Project.Id)
	if err != nil {
		return domain.ServerError("failed to mark project updated at", err)
	}

	return nil
}

func (pas *ProjectActivitySubscriber) handleProjectMemberCreated(ctx context.Context, message Message) error {
	var payload events.ProjectMemberCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal project member created payload", err)
	}

	activity := domain.ProjectMemberCreatedActivity(domain.Project{
		Id: payload.ProjectMember.ProjectId,
	}, payload.ProjectMember, payload.User)
	err = pas.repository.Create(ctx, &activity)
	if err != nil {
		return domain.ServerError("failed to create project member created activity", err)
	}

	err = pas.projectRepository.MarkUpdatedAt(ctx, payload.ProjectMember.ProjectId)
	if err != nil {
		return domain.ServerError("failed to mark project updated at", err)
	}

	return nil
}

func (pas *ProjectActivitySubscriber) handleTaskCreated(ctx context.Context, message Message) error {
	var payload events.TaskCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal task created payload", err)
	}

	activity := domain.TaskCreatedActivity(domain.Project{
		Id: payload.Task.ProjectId,
	}, payload.Task, payload.User)
	err = pas.repository.Create(ctx, &activity)
	if err != nil {
		return domain.ServerError("failed to create task created activity", err)
	}

	err = pas.projectRepository.MarkUpdatedAt(ctx, payload.Task.ProjectId)
	if err != nil {
		return domain.ServerError("failed to mark project updated at", err)
	}

	return nil
}

func (pas *ProjectActivitySubscriber) handleTaskUpdated(ctx context.Context, message Message) error {
	var payload events.TaskUpdatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal task updated payload", err)
	}

	activity := domain.TaskUpdatedActivity(domain.Project{
		Id: payload.Task.ProjectId,
	}, payload.Task, payload.User)
	err = pas.repository.Create(ctx, &activity)
	if err != nil {
		return domain.ServerError("failed to create task updated activity", err)
	}

	err = pas.projectRepository.MarkUpdatedAt(ctx, payload.Task.ProjectId)
	if err != nil {
		return domain.ServerError("failed to mark project updated at", err)
	}

	return nil
}
