package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

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

	pgTypeUuid := pgtype.UUID{Bytes: *chat.ProjectId, Valid: true}

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

func (cr *ChatRepository) DeleteMember(ctx context.Context, member *domain.ChatMember) error {
	q := queries.New(cr.pool)
	return q.DeleteChatMember(ctx, queries.DeleteChatMemberParams{
		UserID: member.UserId,
		ChatID: member.ChatId,
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
	tx, err := cr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(cr.pool)
	qtx := q.WithTx(tx)

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

	id, err := qtx.CreateChatMessage(ctx, params)
	if err != nil {
		return err
	}

	message.Id = id

	err = qtx.UpdateChatUpdatedAt(ctx, queries.UpdateChatUpdatedAtParams{
		UpdatedAt: pgtype.Timestamptz{Time: message.UpdatedAt, Valid: true},
		ID:        message.ChatId,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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

	members, err := unmarshalMembers(chatResult.Members)
	if err != nil {
		return nil, err
	}

	return &domain.Chat{
		Id:        chatResult.ID,
		ProjectId: pgUUIDPointer(chatResult.ProjectID),
		ChatType:  domain.ChatType(chatResult.ChatType),
		CreatedAt: chatResult.CreatedAt.Time,
		UpdatedAt: chatResult.UpdatedAt.Time,
		Members:   members,
		Messages:  []domain.ChatMessage{},
	}, nil
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

	members, err := unmarshalMembers(chatResult.Members)
	if err != nil {
		return nil, err
	}

	return &domain.Chat{
		Id:        chatResult.ID,
		ProjectId: pgUUIDPointer(chatResult.ProjectID),
		ChatType:  domain.ChatType(chatResult.ChatType),
		CreatedAt: chatResult.CreatedAt.Time,
		UpdatedAt: chatResult.UpdatedAt.Time,
		Members:   members,
	}, nil
}

func (cr *ChatRepository) GetOrCreateGeneralChat(ctx context.Context, currentUserId uuid.UUID, memberIds []uuid.UUID) (*domain.Chat, error) {
	q := queries.New(cr.pool)

	existingId, err := q.FindGeneralChatByExactMembers(ctx, memberIds)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return cr.GetById(ctx, existingId)
	}

	tx, err := cr.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	chatId, err := qtx.CreateGeneralChat(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, userId := range memberIds {
		err = qtx.CreateChatMember(ctx, queries.CreateChatMemberParams{
			UserID:     userId,
			ChatID:     chatId,
			LastSeenAt: pgtype.Timestamptz{Time: now, Valid: true},
			JoinedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		})
		if err != nil {
			return nil, err
		}
	}

	chatResult, err := qtx.GetChatById(ctx, chatId)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	members, err := unmarshalMembers(chatResult.Members)
	if err != nil {
		return nil, err
	}

	return &domain.Chat{
		Id:        chatResult.ID,
		ProjectId: pgUUIDPointer(chatResult.ProjectID),
		ChatType:  domain.ChatType(chatResult.ChatType),
		CreatedAt: chatResult.CreatedAt.Time,
		UpdatedAt: chatResult.UpdatedAt.Time,
		Members:   members,
	}, nil
}

func (cr *ChatRepository) ListGeneralChats(ctx context.Context, userId uuid.UUID) ([]domain.Chat, error) {
	q := queries.New(cr.pool)
	rows, err := q.ListGeneralChatsByUserId(ctx, queries.ListGeneralChatsByUserIdParams{
		UserID:  userId,
		Column2: domain.ChatUnreadCountFetchLimit,
	})
	if err != nil {
		return nil, err
	}

	chats := []domain.Chat{}
	for _, row := range rows {
		members, err := unmarshalMembers(row.Members)
		if err != nil {
			return nil, err
		}

		chats = append(chats, domain.Chat{
			Id:            row.ID,
			ProjectId:     pgUUIDPointer(row.ProjectID),
			ChatType:      domain.ChatType(row.ChatType),
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			UnreadCount:   int(row.UnreadCount),
			HasMoreUnread: row.HasMoreUnread,
			Members:       members,
		})
	}

	return chats, nil
}

func (cr *ChatRepository) ListMessages(ctx context.Context, chatId uuid.UUID, params utils.PaginationBeforeParams) ([]domain.ChatMessage, error) {
	q := queries.New(cr.pool)

	queriesParams := queries.ListChatMessagesParams{
		ChatID:    chatId,
		Limit:     params.Limit,
		CreatedAt: pgtype.Timestamptz{Time: params.Before, Valid: true},
		Column3:   params.Id,
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
			ReadsCount:  int(messageResult.ReadsCount),
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

func (cr *ChatRepository) GetUnreadSummary(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (*domain.ChatUnreadSummary, error) {
	q := queries.New(cr.pool)
	row, err := q.GetUnreadCountByChatId(ctx, queries.GetUnreadCountByChatIdParams{
		ChatID:  chatId,
		UserID:  userId,
		Column3: domain.ChatUnreadCountFetchLimit,
	})
	if err != nil {
		return nil, err
	}

	return &domain.ChatUnreadSummary{
		UnreadCount:   int(row.UnreadCount),
		HasMoreUnread: row.HasMoreUnread,
	}, nil
}

func (cr *ChatRepository) GetMessageById(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error) {
	q := queries.New(cr.pool)
	message, err := q.GetChatMessageById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("chat message not found")
		}
		return nil, err
	}

	result := &domain.ChatMessage{
		Id:          message.ID,
		ChatId:      message.ChatID,
		MessageType: domain.MessageType(message.MessageType),
		Content:     message.Content,
		CreatedAt:   message.CreatedAt.Time,
		UpdatedAt:   message.UpdatedAt.Time,
	}

	if message.UserID.Valid {
		userId := uuid.UUID(message.UserID.Bytes)
		result.UserId = &userId
	}

	return result, nil
}

func (cr *ChatRepository) MarkReadUpTo(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, readAt time.Time, message *domain.ChatMessage) error {
	tx, err := cr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(cr.pool)
	qtx := q.WithTx(tx)

	err = qtx.UpdateChatMemberLastSeenAt(ctx, queries.UpdateChatMemberLastSeenAtParams{
		LastSeenAt: pgtype.Timestamptz{Time: readAt, Valid: true},
		UserID:     userId,
		ChatID:     chatId,
	})
	if err != nil {
		return err
	}

	if message != nil {
		err = qtx.UpsertChatMessageReadsUpTo(ctx, queries.UpsertChatMessageReadsUpToParams{
			ChatID:    chatId,
			UserID:    userId,
			ReadAt:    pgtype.Timestamptz{Time: readAt, Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: message.CreatedAt, Valid: true},
			Column5:   message.Id,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (cr *ChatRepository) ListMessageReads(ctx context.Context, messageId uuid.UUID) ([]domain.ChatMessageRead, error) {
	q := queries.New(cr.pool)
	rows, err := q.ListChatMessageReads(ctx, messageId)
	if err != nil {
		return nil, err
	}

	reads := make([]domain.ChatMessageRead, 0, len(rows))
	for _, row := range rows {
		read := domain.ChatMessageRead{
			MessageId: row.ChatMessageID,
			UserId:    row.UserID,
			ReadAt:    row.ReadAt.Time,
			User: &domain.User{
				Id:    row.UserID,
				Name:  row.Name,
				Email: row.Email,
			},
		}

		reads = append(reads, read)
	}

	return reads, nil
}

func unmarshalMembers(raw any) ([]domain.ChatMember, error) {
	if raw == nil {
		return []domain.ChatMember{}, nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var members []domain.ChatMember
	if err := json.Unmarshal(b, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func pgUUIDPointer(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id := uuid.UUID(value.Bytes)
	return &id
}
