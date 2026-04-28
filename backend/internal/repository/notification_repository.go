package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

type notificationRow queries.ListNotificationsByUserIdRow

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{
		pool: pool,
	}
}

func (nr *NotificationRepository) CreateMany(ctx context.Context, notifications []domain.Notification) ([]uuid.UUID, error) {
	if len(notifications) == 0 {
		return []uuid.UUID{}, nil
	}

	tx, err := nr.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := queries.New(nr.pool)
	qtx := q.WithTx(tx)

	ids := make([]uuid.UUID, 0, len(notifications))
	for _, notification := range notifications {
		id, err := qtx.CreateNotification(ctx, queries.CreateNotificationParams{
			UserID:        notification.UserId,
			ActorID:       notification.ActorId,
			ProjectID:     notification.ProjectId,
			TaskID:        optionalUUID(notification.TaskId),
			TaskCommentID: optionalUUID(notification.TaskCommentId),
			Type:          string(notification.Type),
			ReadAt:        optionalTime(notification.ReadAt),
			CreatedAt:     pgtype.Timestamptz{Time: notification.CreatedAt, Valid: true},
			UpdatedAt:     pgtype.Timestamptz{Time: notification.UpdatedAt, Valid: true},
		})
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return ids, nil
}

func (nr *NotificationRepository) ListByUserID(ctx context.Context, userId uuid.UUID, beforeCreatedAt time.Time, beforeId uuid.UUID, limit int32) (*utils.CursorPaginated[domain.Notification], error) {
	q := queries.New(nr.pool)

	rows, err := q.ListNotificationsByUserId(ctx, queries.ListNotificationsByUserIdParams{
		UserID:          userId,
		BeforeCreatedAt: pgtype.Timestamptz{Time: beforeCreatedAt, Valid: !beforeCreatedAt.IsZero()},
		BeforeID:        pgtype.UUID{Bytes: beforeId, Valid: beforeId != uuid.Nil},
		Limit:           pgtype.Int4{Int32: limit, Valid: limit > 0},
	})
	if err != nil {
		return nil, err
	}

	notifications := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		notification, err := mapNotificationRow(notificationRow(row))
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	hasNext := len(notifications) > int(limit)
	data := notifications
	if hasNext {
		data = notifications[:limit]
	}

	return &utils.CursorPaginated[domain.Notification]{
		Data:    data,
		HasNext: hasNext,
	}, nil
}

func (nr *NotificationRepository) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Notification, error) {
	if len(ids) == 0 {
		return []domain.Notification{}, nil
	}

	q := queries.New(nr.pool)
	rows, err := q.ListNotificationsByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	notifications := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		notification, err := mapNotificationRow(notificationRow(row))
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (nr *NotificationRepository) CountUnreadByUserID(ctx context.Context, userId uuid.UUID) (int, error) {
	q := queries.New(nr.pool)
	count, err := q.CountUnreadNotificationsByUserId(ctx, userId)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (nr *NotificationRepository) MarkRead(ctx context.Context, notificationId uuid.UUID, userId uuid.UUID, updatedAt time.Time) (bool, error) {
	q := queries.New(nr.pool)
	rowsAffected, err := q.MarkNotificationRead(ctx, queries.MarkNotificationReadParams{
		ID:        notificationId,
		UserID:    userId,
		UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
	})
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (nr *NotificationRepository) MarkAllRead(ctx context.Context, userId uuid.UUID, updatedAt time.Time) error {
	q := queries.New(nr.pool)
	return q.MarkAllNotificationsRead(ctx, queries.MarkAllNotificationsReadParams{
		UserID:    userId,
		UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
	})
}

func mapNotificationRow(row notificationRow) (domain.Notification, error) {
	notification := domain.Notification{
		Id:        row.ID,
		UserId:    row.UserID,
		ActorId:   row.ActorID,
		ProjectId: row.ProjectID,
		Type:      domain.NotificationType(row.Type),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}

	if row.TaskID.Valid {
		taskID := uuid.UUID(row.TaskID.Bytes)
		notification.TaskId = &taskID
	}

	if row.TaskCommentID.Valid {
		commentID := uuid.UUID(row.TaskCommentID.Bytes)
		notification.TaskCommentId = &commentID
	}

	if row.ReadAt.Valid {
		notification.ReadAt = &row.ReadAt.Time
	}

	if err := json.Unmarshal(row.Actor, &notification.Actor); err != nil {
		return domain.Notification{}, err
	}

	if err := json.Unmarshal(row.Project, &notification.Project); err != nil {
		return domain.Notification{}, err
	}

	if row.Task != nil {
		bytes, err := json.Marshal(row.Task)
		if err != nil {
			return domain.Notification{}, err
		}

		var task domain.Task
		if err := json.Unmarshal(bytes, &task); err != nil {
			return domain.Notification{}, err
		}

		notification.Task = &task
	}

	if row.TaskComment != nil {
		bytes, err := json.Marshal(row.TaskComment)
		if err != nil {
			return domain.Notification{}, err
		}

		var taskComment domain.TaskComment
		if err := json.Unmarshal(bytes, &taskComment); err != nil {
			return domain.Notification{}, err
		}

		notification.TaskComment = &taskComment
	}

	return notification, nil
}

func optionalUUID(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}

	return pgtype.UUID{Bytes: *value, Valid: true}
}

func optionalTime(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{Time: *value, Valid: true}
}
