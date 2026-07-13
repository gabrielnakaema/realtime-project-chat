package notification

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

type Service struct {
	repository notificationRepository
}

func NewService(repository notificationRepository) *Service {
	return &Service{
		repository: repository,
	}
}

type ListRequest struct {
	UserId          uuid.UUID
	BeforeCreatedAt time.Time
	BeforeId        uuid.UUID
	Limit           int32
}

func (ns *Service) List(ctx context.Context, request ListRequest) (*utils.CursorPaginated[domain.Notification], error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	notifications, err := ns.repository.ListByUserID(ctx, request.UserId, request.BeforeCreatedAt, request.BeforeId, request.Limit)
	if err != nil {
		return nil, domain.ServerError("failed to list notifications", err)
	}

	return notifications, nil
}

func (ns *Service) CountUnread(ctx context.Context, userId uuid.UUID) (int, error) {
	if userId == uuid.Nil {
		return 0, domain.UnauthorizedError("unauthorized")
	}

	count, err := ns.repository.CountUnreadByUserID(ctx, userId)
	if err != nil {
		return 0, domain.ServerError("failed to count unread notifications", err)
	}

	return count, nil
}

type MarkReadRequest struct {
	NotificationId uuid.UUID
	UserId         uuid.UUID
}

func (ns *Service) MarkRead(ctx context.Context, request MarkReadRequest) error {
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

func (ns *Service) MarkAllRead(ctx context.Context, userId uuid.UUID) error {
	if userId == uuid.Nil {
		return domain.UnauthorizedError("unauthorized")
	}

	if err := ns.repository.MarkAllRead(ctx, userId, time.Now()); err != nil {
		return domain.ServerError("failed to mark all notifications read", err)
	}

	return nil
}
