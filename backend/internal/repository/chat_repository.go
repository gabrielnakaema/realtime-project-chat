package repository

import (
	"context"
	"errors"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{
		pool: pool,
	}
}

func (cr *ChatRepository) Create(ctx context.Context, chat *domain.Chat) error {
	tx, err := cr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(cr.pool)
	qtx := q.WithTx(tx)

	pgTypeUuid := pgtype.UUID{Bytes: chat.ProjectId, Valid: true}

	id, err := qtx.CreateChat(ctx, pgTypeUuid)
	if err != nil {
		return err
	}

	chat.Id = id

	for i, member := range chat.Members {
		params := queries.CreateChatMemberParams{
			UserID:     member.UserId,
			ChatID:     chat.Id,
			LastSeenAt: pgtype.Timestamptz{Time: member.LastSeenAt, Valid: true},
			JoinedAt:   pgtype.Timestamptz{Time: member.JoinedAt, Valid: true},
		}

		err := qtx.CreateChatMember(ctx, params)
		if err != nil {
			return err
		}

		chat.Members[i].Id = id
	}

	return tx.Commit(ctx)
}

func (cr *ChatRepository) CreateMember(ctx context.Context, member *domain.ChatMember) error {
	q := queries.New(cr.pool)
	return q.CreateChatMember(ctx, queries.CreateChatMemberParams{
		UserID:     member.UserId,
		ChatID:     member.ChatId,
		LastSeenAt: pgtype.Timestamptz{Time: member.LastSeenAt, Valid: true},
		JoinedAt:   pgtype.Timestamptz{Time: member.JoinedAt, Valid: true},
	})
}

func (cr *ChatRepository) UpdateMemberLastSeenAt(ctx context.Context, member *domain.ChatMember) error {
	q := queries.New(cr.pool)
	return q.UpdateChatMemberLastSeenAt(ctx, queries.UpdateChatMemberLastSeenAtParams{
		LastSeenAt: pgtype.Timestamptz{Time: member.LastSeenAt, Valid: true},
		UserID:     member.UserId,
		ChatID:     member.ChatId,
	})
}

func (cr *ChatRepository) CreateMessage(ctx context.Context, message *domain.ChatMessage) error {
	q := queries.New(cr.pool)
	params := queries.CreateChatMessageParams{
		ChatID:      message.ChatId,
		MessageType: string(message.MessageType),
		Content:     message.Content,
		CreatedAt:   pgtype.Timestamptz{Time: message.CreatedAt, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: message.UpdatedAt, Valid: true},
	}

	if message.UserId != nil {
		params.UserID = pgtype.UUID{Bytes: *message.UserId, Valid: true}
	}

	return q.CreateChatMessage(ctx, params)
}

func (cr *ChatRepository) GetByProjectId(ctx context.Context, projectId uuid.UUID) (*domain.Chat, error) {
	q := queries.New(cr.pool)
	chatResult, err := q.GetChatByProjectId(ctx, pgtype.UUID{Bytes: projectId, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("chat not found")
		}
		return nil, err
	}

	chat := domain.Chat{
		Id:        chatResult.ID,
		ProjectId: chatResult.ProjectID.Bytes,
		CreatedAt: chatResult.CreatedAt.Time,
		UpdatedAt: chatResult.UpdatedAt.Time,
		Members:   []domain.ChatMember{},
		Messages:  []domain.ChatMessage{},
	}

	return &chat, nil
}
