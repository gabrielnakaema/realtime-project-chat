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
	CreateMessage(ctx context.Context, message *domain.ChatMessage) error
	UpdateMemberLastSeenAt(ctx context.Context, member *domain.ChatMember) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.Chat, error)
	ListMessages(ctx context.Context, chatId uuid.UUID, params utils.PaginationBeforeParams) ([]domain.ChatMessage, error)
}

type chatUserRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type publisher interface {
	Publish(ctx context.Context, topic events.Topic, payload interface{}) error
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
		ProjectId: project.Id,
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

	err = cs.publisher.Publish(ctx, events.ChatMemberCreated, member)
	if err != nil {
		fmt.Println("failed to publish chat member created event", err)
		return domain.ServerError("failed to publish chat member created event", err)
	}

	return nil
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

	err = cs.publisher.Publish(ctx, events.ChatMessageCreated, message)
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

	err = cs.publisher.Publish(ctx, events.ChatMessageCreated, message)
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

	return chat, nil
}
