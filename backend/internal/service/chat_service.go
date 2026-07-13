package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type chatRepository interface {
	Create(ctx context.Context, chat *domain.Chat) error
	GetByProjectId(ctx context.Context, projectId uuid.UUID) (*domain.Chat, error)
	CreateMember(ctx context.Context, member *domain.ChatMember) error
	DeleteMember(ctx context.Context, member *domain.ChatMember) error
	CreateMessage(ctx context.Context, message *domain.ChatMessage) error
	UpdateMemberLastSeenAt(ctx context.Context, member *domain.ChatMember) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.Chat, error)
	ListMessages(ctx context.Context, chatId uuid.UUID, params utils.PaginationBeforeParams) ([]domain.ChatMessage, error)
	GetOrCreateGeneralChat(ctx context.Context, currentUserId uuid.UUID, memberIds []uuid.UUID) (*domain.Chat, error)
	ListGeneralChats(ctx context.Context, userId uuid.UUID) ([]domain.Chat, error)
	GetUnreadSummary(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (*domain.ChatUnreadSummary, error)
	GetMessageById(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error)
	MarkReadUpTo(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, readAt time.Time, message *domain.ChatMessage) error
	ListMessageReads(ctx context.Context, messageId uuid.UUID) ([]domain.ChatMessageRead, error)
}

type chatUserRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type publisher interface {
	Publish(ctx context.Context, topic events.Topic, payload events.Payload) error
}

type ChatService struct {
	chatRepository chatRepository
	userRepository chatUserRepository
	publisher      publisher
}

func NewChatService(chatRepository chatRepository, userRepository chatUserRepository, publisher publisher) *ChatService {
	return &ChatService{
		chatRepository: chatRepository,
		userRepository: userRepository,
		publisher:      publisher,
	}
}

func (cs *ChatService) CreateChatFromProject(ctx context.Context, project *domain.Project) error {
	members := []domain.ChatMember{}
	for _, member := range project.Members {
		members = append(members, domain.ChatMember{
			UserId:     member.UserId,
			ChatId:     project.Id,
			JoinedAt:   time.Now(),
			LastSeenAt: time.Now(),
		})
	}

	chat := domain.Chat{
		ProjectId: &project.Id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Members:   members,
		Messages:  []domain.ChatMessage{},
	}

	err := cs.chatRepository.Create(ctx, &chat)
	if err != nil {
		return domain.ServerError("failed to create chat", err)
	}

	return nil
}

func (cs *ChatService) CreateMemberFromProjectMember(ctx context.Context, projectMember *domain.ProjectMember) error {
	chat, err := cs.chatRepository.GetByProjectId(ctx, projectMember.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			return err
		}
		return domain.ServerError("failed to get chat", err)
	}

	member := domain.ChatMember{
		UserId:     projectMember.UserId,
		ChatId:     chat.Id,
		JoinedAt:   time.Now(),
		LastSeenAt: time.Now(),
	}

	err = cs.chatRepository.CreateMember(ctx, &member)
	if err != nil {
		return domain.ServerError("failed to create member", err)
	}

	err = cs.publisher.Publish(ctx, events.ChatMemberCreated, &events.ChatMemberCreatedPayload{
		ChatMember:    member,
		ProjectMember: *projectMember,
		User: domain.User{
			Id: projectMember.UserId,
		},
	})
	if err != nil {
		return domain.ServerError("failed to publish chat member created event", err)
	}

	return nil
}

func (cs *ChatService) RemoveMemberFromProjectMember(ctx context.Context, projectMember *domain.ProjectMember) (*domain.ChatMember, error) {
	chat, err := cs.chatRepository.GetByProjectId(ctx, projectMember.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project chat", err)
	}

	member := &domain.ChatMember{
		UserId: projectMember.UserId,
		ChatId: chat.Id,
	}
	if err := cs.chatRepository.DeleteMember(ctx, member); err != nil {
		return nil, domain.ServerError("failed to remove project chat member", err)
	}

	if err := cs.publisher.Publish(ctx, events.ChatMemberRemoved, &events.ChatMemberRemovedPayload{
		ChatMember:    *member,
		ProjectMember: *projectMember,
		User:          domain.User{Id: projectMember.UserId},
	}); err != nil {
		return nil, domain.ServerError("failed to publish chat member removed event", err)
	}

	return member, nil
}

func (cs *ChatService) CreateJoinedMessage(ctx context.Context, chatMember *domain.ChatMember) error {
	user, err := cs.userRepository.GetById(ctx, chatMember.UserId)
	if err != nil {
		return domain.ServerError("failed to get user", err)
	}

	message := domain.ChatMessage{
		ChatId:      chatMember.ChatId,
		MessageType: domain.MessageTypeSystem,
		UserId:      nil,
		Content:     fmt.Sprintf("%s has joined the chat", user.Name),
		CreatedAt:   chatMember.JoinedAt,
		UpdatedAt:   chatMember.JoinedAt,
	}

	err = cs.chatRepository.CreateMessage(ctx, &message)
	if err != nil {
		return domain.ServerError("failed to create joined message", err)
	}

	err = cs.publisher.Publish(ctx, events.ChatMessageCreated, &events.ChatMessageCreatedPayload{
		ChatMessage: message,
		User: domain.User{
			Id: chatMember.UserId,
		},
	})
	if err != nil {
		return domain.ServerError("failed to create publisher event", err)
	}

	return nil
}

type CreateChatMessageRequest struct {
	ChatId  uuid.UUID
	UserId  uuid.UUID
	Content string
}

func (cs *ChatService) CreateMessage(ctx context.Context, request CreateChatMessageRequest) (*domain.ChatMessage, error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	chat, err := cs.chatRepository.GetById(ctx, request.ChatId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("chat not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get chat", err)
	}

	var foundMember *domain.ChatMember
	hasPermission := false
	for _, member := range chat.Members {
		if member.UserId == request.UserId {
			hasPermission = true
			foundMember = &member
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	message := domain.ChatMessage{
		MessageType: domain.MessageTypeText,
		Member:      foundMember,
		ChatId:      request.ChatId,
		UserId:      &request.UserId,
		Content:     request.Content,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = cs.chatRepository.CreateMessage(ctx, &message)
	if err != nil {
		return nil, domain.ServerError("failed to create message", err)
	}

	err = cs.publisher.Publish(ctx, events.ChatMessageCreated, &events.ChatMessageCreatedPayload{
		ChatMessage: message,
		User: domain.User{
			Id: request.UserId,
		},
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish chat message created event", err)
	}

	return &message, nil
}

func (cs *ChatService) GetByProjectId(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*domain.Chat, error) {
	if userId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	chat, err := cs.chatRepository.GetByProjectId(ctx, projectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("chat not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get chat", err)
	}

	hasPermission := false
	for _, member := range chat.Members {
		if member.UserId == userId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	unreadSummary, err := cs.chatRepository.GetUnreadSummary(ctx, chat.Id, userId)
	if err != nil {
		return nil, domain.ServerError("failed to count unread messages", err)
	}
	chat.UnreadCount = unreadSummary.UnreadCount
	chat.HasMoreUnread = unreadSummary.HasMoreUnread

	return chat, nil
}

type ListMessagesByProjectIdRequest struct {
	ProjectId uuid.UUID
	UserId    uuid.UUID
	Params    utils.PaginationBeforeParams
}

func (cs *ChatService) ListMessagesByProjectId(ctx context.Context, request ListMessagesByProjectIdRequest) (*utils.CursorPaginated[domain.ChatMessage], error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	chat, err := cs.chatRepository.GetByProjectId(ctx, request.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get chat", err)
	}

	hasPermission := false
	for _, member := range chat.Members {
		if member.UserId == request.UserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	chat.Messages, err = cs.chatRepository.ListMessages(ctx, chat.Id, request.Params)
	if err != nil {
		return nil, domain.ServerError("failed to list messages", err)
	}

	slices.Reverse(chat.Messages)

	cursorPaginated := utils.CursorPaginated[domain.ChatMessage]{
		Data:    chat.Messages,
		HasNext: len(chat.Messages) >= int(request.Params.Limit),
	}

	return &cursorPaginated, nil
}

func (cs *ChatService) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Chat, error) {
	if userId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	chat, err := cs.chatRepository.GetById(ctx, id)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("chat not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get chat", err)
	}

	hasPermission := false
	for _, member := range chat.Members {
		if member.UserId == userId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	unreadSummary, err := cs.chatRepository.GetUnreadSummary(ctx, chat.Id, userId)
	if err != nil {
		return nil, domain.ServerError("failed to count unread messages", err)
	}
	chat.UnreadCount = unreadSummary.UnreadCount
	chat.HasMoreUnread = unreadSummary.HasMoreUnread

	return chat, nil
}

func (cs *ChatService) UpdateMemberLastSeenAt(ctx context.Context, userId uuid.UUID, chatId uuid.UUID) error {
	if userId == uuid.Nil {
		return domain.UnauthorizedError("unauthorized")
	}

	chatMember := &domain.ChatMember{
		UserId:     userId,
		ChatId:     chatId,
		LastSeenAt: time.Now(),
	}

	err := cs.chatRepository.UpdateMemberLastSeenAt(ctx, chatMember)
	if err != nil {
		return err
	}

	return nil
}

func SanitizeGeneralChatMemberIDs(currentUserId uuid.UUID, targetUserIds []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	memberIds := []uuid.UUID{currentUserId}
	seen[currentUserId] = struct{}{}

	for _, userId := range targetUserIds {
		if userId == uuid.Nil || userId == currentUserId {
			continue
		}

		if _, ok := seen[userId]; ok {
			continue
		}

		seen[userId] = struct{}{}
		memberIds = append(memberIds, userId)
	}

	return memberIds
}

func (cs *ChatService) GetOrCreateGeneralChat(ctx context.Context, currentUserId uuid.UUID, targetUserIds []uuid.UUID) (*domain.Chat, error) {
	if currentUserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	memberIds := SanitizeGeneralChatMemberIDs(currentUserId, targetUserIds)
	if len(memberIds) <= 1 {
		return nil, domain.BusinessValidationError("at least one other user is required")
	}

	for _, memberId := range memberIds[1:] {
		_, err := cs.userRepository.GetById(ctx, memberId)
		if err != nil {
			var domainErr domain.DomainError
			if errors.As(err, &domainErr) {
				if domainErr.Code == domain.NotFoundErrorCode {
					return nil, domain.NotFoundError("user not found")
				}
				return nil, domainErr
			}
			return nil, domain.ServerError("failed to get user", err)
		}
	}

	chat, err := cs.chatRepository.GetOrCreateGeneralChat(ctx, currentUserId, memberIds)
	if err != nil {
		return nil, domain.ServerError("failed to get or create general chat", err)
	}

	return chat, nil
}

func (cs *ChatService) ListGeneralChats(ctx context.Context, userId uuid.UUID) ([]domain.Chat, error) {
	if userId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	chats, err := cs.chatRepository.ListGeneralChats(ctx, userId)
	if err != nil {
		return nil, domain.ServerError("failed to list general chats", err)
	}

	return chats, nil
}

type ListMessagesByChatIdRequest struct {
	ChatId uuid.UUID
	UserId uuid.UUID
	Params utils.PaginationBeforeParams
}

func (cs *ChatService) ListMessagesByChatId(ctx context.Context, request ListMessagesByChatIdRequest) (*utils.CursorPaginated[domain.ChatMessage], error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	chat, err := cs.chatRepository.GetById(ctx, request.ChatId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("chat not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get chat", err)
	}

	hasPermission := false
	for _, member := range chat.Members {
		if member.UserId == request.UserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	messages, err := cs.chatRepository.ListMessages(ctx, chat.Id, request.Params)
	if err != nil {
		return nil, domain.ServerError("failed to list messages", err)
	}

	slices.Reverse(messages)

	return &utils.CursorPaginated[domain.ChatMessage]{
		Data:    messages,
		HasNext: len(messages) >= int(request.Params.Limit),
	}, nil
}

type MarkChatReadRequest struct {
	ChatId    uuid.UUID
	UserId    uuid.UUID
	MessageId *uuid.UUID
}

func (cs *ChatService) MarkChatRead(ctx context.Context, request MarkChatReadRequest) error {
	if request.UserId == uuid.Nil {
		return domain.UnauthorizedError("unauthorized")
	}

	chat, err := cs.chatRepository.GetById(ctx, request.ChatId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			return domainErr
		}
		return domain.ServerError("failed to get chat", err)
	}

	hasPermission := false
	for _, member := range chat.Members {
		if member.UserId == request.UserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return domain.ForbiddenError("forbidden")
	}

	var message *domain.ChatMessage
	if request.MessageId != nil {
		message, err = cs.chatRepository.GetMessageById(ctx, *request.MessageId)
		if err != nil {
			var domainErr domain.DomainError
			if errors.As(err, &domainErr) {
				return domainErr
			}
			return domain.ServerError("failed to get chat message", err)
		}

		if message.ChatId != request.ChatId {
			return domain.ForbiddenError("forbidden")
		}
	}

	readAt := time.Now()
	err = cs.chatRepository.MarkReadUpTo(ctx, request.ChatId, request.UserId, readAt, message)
	if err != nil {
		return domain.ServerError("failed to mark chat as read", err)
	}

	if message != nil {
		err = cs.publisher.Publish(ctx, events.ChatMessageRead, &events.ChatMessageReadPayload{
			ChatID:    request.ChatId,
			MessageID: message.Id,
			Read: domain.ChatMessageRead{
				MessageId: message.Id,
				UserId:    request.UserId,
				ReadAt:    readAt,
			},
		})
		if err != nil {
			return domain.ServerError("failed to publish chat read event", err)
		}
	}

	return nil
}

type ListMessageReadsRequest struct {
	ChatId    uuid.UUID
	MessageId uuid.UUID
	UserId    uuid.UUID
}

func (cs *ChatService) ListMessageReads(ctx context.Context, request ListMessageReadsRequest) ([]domain.ChatMessageRead, error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	chat, err := cs.chatRepository.GetById(ctx, request.ChatId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get chat", err)
	}

	hasPermission := false
	for _, member := range chat.Members {
		if member.UserId == request.UserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	message, err := cs.chatRepository.GetMessageById(ctx, request.MessageId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get chat message", err)
	}
	if message.ChatId != request.ChatId {
		return nil, domain.ForbiddenError("forbidden")
	}

	reads, err := cs.chatRepository.ListMessageReads(ctx, request.MessageId)
	if err != nil {
		return nil, domain.ServerError("failed to list message reads", err)
	}

	return reads, nil
}
