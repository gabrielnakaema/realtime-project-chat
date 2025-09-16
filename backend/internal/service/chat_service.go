package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/google/uuid"
)

type chatRepository interface {
	Create(ctx context.Context, chat *domain.Chat) error
	GetByProjectId(ctx context.Context, projectId uuid.UUID) (*domain.Chat, error)
	CreateMember(ctx context.Context, member *domain.ChatMember) error
	CreateMessage(ctx context.Context, message *domain.ChatMessage) error
	UpdateMemberLastSeenAt(ctx context.Context, member *domain.ChatMember) error
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

	return nil
}
