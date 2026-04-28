package service

import (
	"context"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type notificationRepository interface {
	ListByUserID(ctx context.Context, userId uuid.UUID, beforeCreatedAt time.Time, beforeId uuid.UUID, limit int32) (*utils.CursorPaginated[domain.Notification], error)
	CountUnreadByUserID(ctx context.Context, userId uuid.UUID) (int, error)
	MarkRead(ctx context.Context, notificationId uuid.UUID, userId uuid.UUID, updatedAt time.Time) (bool, error)
	MarkAllRead(ctx context.Context, userId uuid.UUID, updatedAt time.Time) error
}

type NotificationService struct {
	repository notificationRepository
}

func NewNotificationService(repository notificationRepository) *NotificationService {
	return &NotificationService{
		repository: repository,
	}
}

type ListNotificationsRequest struct {
	UserId          uuid.UUID
	BeforeCreatedAt time.Time
	BeforeId        uuid.UUID
	Limit           int32
}

func (ns *NotificationService) List(ctx context.Context, request ListNotificationsRequest) (*utils.CursorPaginated[domain.Notification], error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	notifications, err := ns.repository.ListByUserID(ctx, request.UserId, request.BeforeCreatedAt, request.BeforeId, request.Limit)
	if err != nil {
		return nil, domain.ServerError("failed to list notifications", err)
	}

	return notifications, nil
}

func (ns *NotificationService) CountUnread(ctx context.Context, userId uuid.UUID) (int, error) {
	if userId == uuid.Nil {
		return 0, domain.UnauthorizedError("unauthorized")
	}

	count, err := ns.repository.CountUnreadByUserID(ctx, userId)
	if err != nil {
		return 0, domain.ServerError("failed to count unread notifications", err)
	}

	return count, nil
}

type MarkNotificationReadRequest struct {
	NotificationId uuid.UUID
	UserId         uuid.UUID
}

func (ns *NotificationService) MarkRead(ctx context.Context, request MarkNotificationReadRequest) error {
	if request.UserId == uuid.Nil {
		return domain.UnauthorizedError("unauthorized")
	}

	found, err := ns.repository.MarkRead(ctx, request.NotificationId, request.UserId, time.Now())
	if err != nil {
		return domain.ServerError("failed to mark notification read", err)
	}

	if !found {
		return domain.NotFoundError("notification not found")
	}

	return nil
}

func (ns *NotificationService) MarkAllRead(ctx context.Context, userId uuid.UUID) error {
	if userId == uuid.Nil {
		return domain.UnauthorizedError("unauthorized")
	}

	if err := ns.repository.MarkAllRead(ctx, userId, time.Now()); err != nil {
		return domain.ServerError("failed to mark all notifications read", err)
	}

	return nil
}
