package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/gabrielnakaema/project-chat/internal/utils"
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

	for _, member := range chat.Members {
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

	if chatResult.Members != nil {
		bytes, err := json.Marshal(chatResult.Members)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &chat.Members)
		if err != nil {
			return nil, err
		}
	}

	return &chat, nil
}

func (cr *ChatRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Chat, error) {
	q := queries.New(cr.pool)
	chatResult, err := q.GetChatById(ctx, id)
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
	}

	if chatResult.Members != nil {
		bytes, err := json.Marshal(chatResult.Members)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &chat.Members)
		if err != nil {
			return nil, err
		}
	}

	return &chat, nil
}

func (cr *ChatRepository) ListMessages(ctx context.Context, chatId uuid.UUID, params utils.PaginationBeforeParams) ([]domain.ChatMessage, error) {
	q := queries.New(cr.pool)

	queriesParams := queries.ListChatMessagesParams{
		ChatID: chatId,
		Limit:  params.Limit,
	}
	if params.Before != nil {
		queriesParams.CreatedAt = pgtype.Timestamptz{Time: *params.Before, Valid: true}
	}

	result, err := q.ListChatMessages(ctx, queriesParams)
	if err != nil {
		return nil, err
	}

	messages := []domain.ChatMessage{}
	for _, messageResult := range result {
		message := domain.ChatMessage{
			Id:          messageResult.ID,
			ChatId:      messageResult.ChatID,
			MessageType: domain.MessageType(messageResult.MessageType),
			Content:     messageResult.Content,
			CreatedAt:   messageResult.CreatedAt.Time,
			UpdatedAt:   messageResult.UpdatedAt.Time,
		}

		if messageResult.UserID.Valid {
			message.UserId = (*uuid.UUID)(messageResult.UserID.Bytes[:])

			user := domain.User{
				Id:   *message.UserId,
				Name: messageResult.UserName.String,
			}

			message.Member = &domain.ChatMember{
				UserId: *message.UserId,
				ChatId: message.ChatId,
				User:   &user,
			}
		}

		messages = append(messages, message)
	}

	return messages, nil
}
