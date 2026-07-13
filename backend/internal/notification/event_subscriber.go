package notification

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/subscriber"
	"github.com/google/uuid"
)

type eventRepository interface {
	CreateMany(ctx context.Context, notifications []domain.Notification) ([]uuid.UUID, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Notification, error)
}

type Notifier interface {
	SendNotification(ctx context.Context, notification *domain.Notification) error
}

type EventSubscriber struct {
	logger     *slog.Logger
	subscriber *subscriber.Subscriber
	repository eventRepository
	notifier   Notifier
	now        func() time.Time
}

func NewEventSubscriber(ctx context.Context, cfg *config.Config, logger *slog.Logger, repository eventRepository, notifier Notifier) (*EventSubscriber, error) {
	sub, err := subscriber.NewSubscriber(cfg, "notification.subscriber")
	if err != nil {
		return nil, err
	}

	eventSubscriber := &EventSubscriber{
		logger:     logger,
		subscriber: sub,
		repository: repository,
		notifier:   notifier,
		now:        time.Now,
	}

	topics := []events.Topic{events.ProjectMemberCreated, events.TaskCreated, events.TaskUpdated, events.TaskCommentCreated}
	if err := sub.Subscribe(ctx, topics, eventSubscriber.handleNotificationEvents, eventSubscriber.logger); err != nil {
		return nil, err
	}

	return eventSubscriber, nil
}

func (es *EventSubscriber) Close() error {
	return es.subscriber.Close()
}

func (es *EventSubscriber) handleNotificationEvents(ctx context.Context, message subscriber.Message) error {
	switch message.Topic {
	case events.ProjectMemberCreated:
		return es.handleProjectMemberCreated(ctx, message)
	case events.TaskCreated:
		return es.handleTaskCreated(ctx, message)
	case events.TaskUpdated:
		return es.handleTaskUpdated(ctx, message)
	case events.TaskCommentCreated:
		return es.handleTaskCommentCreated(ctx, message)
	default:
		return nil
	}
}

func (es *EventSubscriber) handleProjectMemberCreated(ctx context.Context, message subscriber.Message) error {
	var payload events.ProjectMemberCreatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ServerError("failed to unmarshal project member created payload", err)
	}

	if payload.ProjectMember.UserId == payload.User.Id {
		return nil
	}

	notifications := []domain.Notification{
		{
			UserId:    payload.ProjectMember.UserId,
			ActorId:   payload.User.Id,
			ProjectId: payload.ProjectMember.ProjectId,
			Type:      domain.NotificationTypeProjectMemberCreated,
			CreatedAt: es.now(),
			UpdatedAt: es.now(),
		},
	}

	return es.persistAndNotify(ctx, notifications)
}

func (es *EventSubscriber) handleTaskCreated(ctx context.Context, message subscriber.Message) error {
	var payload events.TaskCreatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ServerError("failed to unmarshal task created payload", err)
	}

	if payload.Task.ResponsibleId == nil || *payload.Task.ResponsibleId == payload.User.Id {
		return nil
	}

	notifications := []domain.Notification{
		{
			UserId:    *payload.Task.ResponsibleId,
			ActorId:   payload.User.Id,
			ProjectId: payload.Task.ProjectId,
			TaskId:    pointerUUID(payload.Task.Id),
			Type:      domain.NotificationTypeTaskAssigned,
			CreatedAt: es.now(),
			UpdatedAt: es.now(),
		},
	}

	return es.persistAndNotify(ctx, notifications)
}

func (es *EventSubscriber) handleTaskUpdated(ctx context.Context, message subscriber.Message) error {
	var payload events.TaskUpdatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ServerError("failed to unmarshal task updated payload", err)
	}

	if payload.PreviousTask == nil || payload.Task.ResponsibleId == nil {
		return nil
	}

	if payload.PreviousTask.ResponsibleId != nil && *payload.PreviousTask.ResponsibleId == *payload.Task.ResponsibleId {
		return nil
	}

	if *payload.Task.ResponsibleId == payload.User.Id {
		return nil
	}

	notifications := []domain.Notification{
		{
			UserId:    *payload.Task.ResponsibleId,
			ActorId:   payload.User.Id,
			ProjectId: payload.Task.ProjectId,
			TaskId:    pointerUUID(payload.Task.Id),
			Type:      domain.NotificationTypeTaskAssigned,
			CreatedAt: es.now(),
			UpdatedAt: es.now(),
		},
	}

	return es.persistAndNotify(ctx, notifications)
}

func (es *EventSubscriber) handleTaskCommentCreated(ctx context.Context, message subscriber.Message) error {
	var payload events.TaskCommentCreatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ServerError("failed to unmarshal task comment created payload", err)
	}

	if payload.TaskComment.Task == nil || payload.TaskComment.User == nil {
		return nil
	}

	recipients := map[uuid.UUID]struct{}{}

	if payload.TaskComment.Task.AuthorId != payload.User.Id {
		recipients[payload.TaskComment.Task.AuthorId] = struct{}{}
	}

	if payload.TaskComment.Task.ResponsibleId != nil && *payload.TaskComment.Task.ResponsibleId != payload.User.Id {
		recipients[*payload.TaskComment.Task.ResponsibleId] = struct{}{}
	}

	if len(recipients) == 0 {
		return nil
	}

	notifications := make([]domain.Notification, 0, len(recipients))
	for recipientId := range recipients {
		notifications = append(notifications, domain.Notification{
			UserId:        recipientId,
			ActorId:       payload.User.Id,
			ProjectId:     payload.TaskComment.Task.ProjectId,
			TaskId:        pointerUUID(payload.TaskComment.Task.Id),
			TaskCommentId: parseOptionalUUID(payload.TaskComment.ID),
			Type:          domain.NotificationTypeTaskCommentCreated,
			CreatedAt:     es.now(),
			UpdatedAt:     es.now(),
		})
	}

	return es.persistAndNotify(ctx, notifications)
}

func (es *EventSubscriber) persistAndNotify(ctx context.Context, notifications []domain.Notification) error {
	ids, err := es.repository.CreateMany(ctx, notifications)
	if err != nil {
		return domain.ServerError("failed to create notifications", err)
	}

	createdNotifications, err := es.repository.ListByIDs(ctx, ids)
	if err != nil {
		return domain.ServerError("failed to list created notifications", err)
	}

	for _, notification := range createdNotifications {
		if err := es.notifier.SendNotification(ctx, &notification); err != nil {
			return domain.ServerError("failed to send notification", err)
		}
	}

	return nil
}

func pointerUUID(id uuid.UUID) *uuid.UUID {
	return &id
}

func parseOptionalUUID(value string) *uuid.UUID {
	if value == "" {
		return nil
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return nil
	}

	return &id
}
