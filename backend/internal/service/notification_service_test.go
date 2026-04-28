package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type notificationRepositoryStub struct {
	listResult        *utils.CursorPaginated[domain.Notification]
	listErr           error
	countUnreadResult int
	countUnreadErr    error
	markReadFound     bool
	markReadErr       error
	markAllReadErr    error
}

func (s *notificationRepositoryStub) ListByUserID(context.Context, uuid.UUID, time.Time, uuid.UUID, int32) (*utils.CursorPaginated[domain.Notification], error) {
	return s.listResult, s.listErr
}

func (s *notificationRepositoryStub) CountUnreadByUserID(context.Context, uuid.UUID) (int, error) {
	return s.countUnreadResult, s.countUnreadErr
}

func (s *notificationRepositoryStub) MarkRead(context.Context, uuid.UUID, uuid.UUID, time.Time) (bool, error) {
	return s.markReadFound, s.markReadErr
}

func (s *notificationRepositoryStub) MarkAllRead(context.Context, uuid.UUID, time.Time) error {
	return s.markAllReadErr
}

func TestNotificationService_ListUnauthorized(t *testing.T) {
	svc := service.NewNotificationService(&notificationRepositoryStub{})

	result, err := svc.List(context.Background(), service.ListNotificationsRequest{
		UserId: uuid.Nil,
	})

	require.Error(t, err)
	assert.Nil(t, result)
	var domainErr domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domain.UnauthorizedErrorCode, domainErr.Code)
}

func TestNotificationService_CountUnreadSuccess(t *testing.T) {
	svc := service.NewNotificationService(&notificationRepositoryStub{
		countUnreadResult: 3,
	})

	count, err := svc.CountUnread(context.Background(), uuid.New())

	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestNotificationService_MarkReadReturnsNotFoundWhenMissing(t *testing.T) {
	svc := service.NewNotificationService(&notificationRepositoryStub{
		markReadFound: false,
	})

	err := svc.MarkRead(context.Background(), service.MarkNotificationReadRequest{
		NotificationId: uuid.New(),
		UserId:         uuid.New(),
	})

	require.Error(t, err)
	var domainErr domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domain.NotFoundErrorCode, domainErr.Code)
}

func TestNotificationService_MarkAllReadWrapsRepositoryError(t *testing.T) {
	svc := service.NewNotificationService(&notificationRepositoryStub{
		markAllReadErr: errors.New("db error"),
	})

	err := svc.MarkAllRead(context.Background(), uuid.New())

	require.Error(t, err)
	var domainErr domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domain.ServerErrorCode, domainErr.Code)
}
