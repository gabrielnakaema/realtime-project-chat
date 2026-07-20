package tasks

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/config"
	"github.com/gabrielnakaema/project-chat/internal/platform/messaging"
	"github.com/google/uuid"
)

// TaskMemberCleanupRepository clears task responsibility when a project member
// is removed.
type TaskMemberCleanupRepository interface {
	ClearResponsibleForProjectMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error
}

// MemberCleanupSubscriber reacts to project member removals by unassigning the
// removed member from any tasks they were responsible for. This runs
// asynchronously; the un-assignment is eventually consistent with the removal.
type MemberCleanupSubscriber struct {
	logger     *slog.Logger
	subscriber *messaging.Subscriber
	repository TaskMemberCleanupRepository
}

func NewMemberCleanupSubscriber(ctx context.Context, config *config.Config, logger *slog.Logger, repository TaskMemberCleanupRepository) (*MemberCleanupSubscriber, error) {
	sub, err := messaging.NewSubscriber(config, "task_member_cleanup.subscriber")
	if err != nil {
		return nil, err
	}

	memberCleanupSubscriber := &MemberCleanupSubscriber{
		logger:     logger,
		subscriber: sub,
		repository: repository,
	}

	topics := []events.Topic{events.ProjectMemberRemoved}
	err = sub.Subscribe(ctx, topics, memberCleanupSubscriber.handleProjectMemberRemoved, memberCleanupSubscriber.logger)
	if err != nil {
		return nil, err
	}

	return memberCleanupSubscriber, nil
}

func (s *MemberCleanupSubscriber) Close() error {
	return s.subscriber.Close()
}

func (s *MemberCleanupSubscriber) handleProjectMemberRemoved(ctx context.Context, message messaging.Message) error {
	var payload events.ProjectMemberRemovedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return apperr.ServerError("failed to unmarshal project member removed payload", err)
	}

	if err := s.repository.ClearResponsibleForProjectMember(ctx, payload.ProjectMember.ProjectId, payload.ProjectMember.UserId); err != nil {
		return apperr.ServerError("failed to clear task responsibility for removed project member", err)
	}

	return nil
}
