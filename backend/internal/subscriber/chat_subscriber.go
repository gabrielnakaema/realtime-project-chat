package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/google/uuid"
)

type MessageNotifier interface {
	SendMessages(ctx context.Context, message *domain.ChatMessage) error
	SendReadUpdate(ctx context.Context, chatId uuid.UUID, read *domain.ChatMessageRead) error
	SendChatProjectMemberCreated(ctx context.Context, member *domain.ProjectMember, chatID uuid.UUID) error
	SendChatProjectMemberRemoved(ctx context.Context, member *domain.ProjectMember, chatID uuid.UUID) error
}

type ChatSubscriber struct {
	logger      *slog.Logger
	subscriber  *Subscriber
	chatService *service.ChatService
	notifier    MessageNotifier
}

func NewChatSubscriber(ctx context.Context, config *config.Config, logger *slog.Logger, chatService *service.ChatService, notifier MessageNotifier) (*ChatSubscriber, error) {
	subscriber, err := NewSubscriber(config, "chat.subscriber")
	if err != nil {
		return nil, domain.ServerError("failed to create chat subscriber", err)
	}

	chatSubscriber := &ChatSubscriber{
		subscriber:  subscriber,
		logger:      logger,
		chatService: chatService,
		notifier:    notifier,
	}

	topics := []events.Topic{events.ProjectCreated, events.ProjectMemberCreated, events.ProjectMemberRemoved, events.ChatMemberCreated, events.ChatMessageCreated, events.ChatMemberViewed, events.ChatMessageRead}

	err = subscriber.Subscribe(ctx, topics, chatSubscriber.handleChatEvents, chatSubscriber.logger)
	if err != nil {
		return nil, domain.ServerError("failed to subscribe to chat events", err)
	}

	return chatSubscriber, nil
}

func (cs *ChatSubscriber) Close() error {
	return cs.subscriber.Close()
}

func (cs *ChatSubscriber) handleChatEvents(ctx context.Context, message Message) error {
	switch message.Topic {
	case events.ProjectCreated:
		return cs.handleProjectCreated(ctx, message)
	case events.ProjectMemberCreated:
		return cs.handleProjectMemberCreated(ctx, message)
	case events.ProjectMemberRemoved:
		return cs.handleProjectMemberRemoved(ctx, message)
	case events.ChatMemberCreated:
		return cs.handleChatMemberCreated(ctx, message)
	case events.ChatMessageCreated:
		return cs.handleChatMessageCreated(ctx, message)
	case events.ChatMemberViewed:
		return cs.handleChatMemberViewed(ctx, message)
	case events.ChatMessageRead:
		return cs.handleChatMessageRead(ctx, message)
	}

	return nil
}

func (cs *ChatSubscriber) handleChatMemberViewed(ctx context.Context, message Message) error {
	var payload events.ChatMemberViewedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal chat member", err)
	}

	err = cs.chatService.UpdateMemberLastSeenAt(ctx, payload.ChatMember.UserId, payload.ChatMember.ChatId)
	if err != nil {
		return err
	}

	return nil
}

func (cs *ChatSubscriber) handleProjectCreated(ctx context.Context, message Message) error {
	var payload events.ProjectCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal project", err)
	}

	err = cs.chatService.CreateChatFromProject(ctx, &payload.Project)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.ServerErrorCode && domainErr.Cause != nil {
				cs.logger.Error("failed to create chat from project", "error", domainErr.Cause.Error())
				return nil
			}
			return err
		}
		cs.logger.Error("failed to create chat from project", "error", err.Error())
		return domain.ServerError("failed to create chat from project", err)
	}

	return nil
}

func (cs *ChatSubscriber) handleProjectMemberCreated(ctx context.Context, message Message) error {
	var payload events.ProjectMemberCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal project member", err)
	}

	err = cs.chatService.CreateMemberFromProjectMember(ctx, &payload.ProjectMember)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				cs.logger.Info("chat not found, skipping creation of chat member from project member", "project_member", payload.ProjectMember)
				return nil
			}
			return err
		}

		cs.logger.Error("failed to create member from project member", "error", err)
		return domain.ServerError("failed to create member from project member", err)
	}

	chat, err := cs.chatService.GetByProjectId(ctx, payload.ProjectMember.ProjectId, payload.ProjectMember.UserId)
	if err != nil {
		return err
	}
	if err := cs.notifier.SendChatProjectMemberCreated(ctx, &payload.ProjectMember, chat.Id); err != nil {
		return domain.ServerError("failed to send created project member to chat room", err)
	}

	return nil
}

func (cs *ChatSubscriber) handleProjectMemberRemoved(ctx context.Context, message Message) error {
	var payload events.ProjectMemberRemovedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ServerError("failed to unmarshal project member removed payload", err)
	}

	chatMember, err := cs.chatService.RemoveMemberFromProjectMember(ctx, &payload.ProjectMember)
	if err != nil {
		return err
	}
	if err := cs.notifier.SendChatProjectMemberRemoved(ctx, &payload.ProjectMember, chatMember.ChatId); err != nil {
		return domain.ServerError("failed to send removed project member to chat room", err)
	}

	return nil
}

func (cs *ChatSubscriber) handleChatMemberCreated(ctx context.Context, message Message) error {
	var payload events.ChatMemberCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal chat member", err)
	}

	err = cs.chatService.CreateJoinedMessage(ctx, &payload.ChatMember)
	if err != nil {
		return domain.ServerError("failed to update member last seen at", err)
	}

	return nil
}

func (cs *ChatSubscriber) handleChatMessageCreated(ctx context.Context, message Message) error {
	var payload events.ChatMessageCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal chat message", err)
	}

	err = cs.notifier.SendMessages(ctx, &payload.ChatMessage)
	if err != nil {
		return domain.ServerError("failed to send messages", err)
	}

	return nil
}

func (cs *ChatSubscriber) handleChatMessageRead(ctx context.Context, message Message) error {
	var payload events.ChatMessageReadPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return domain.ServerError("failed to unmarshal chat message read", err)
	}

	return cs.notifier.SendReadUpdate(ctx, payload.ChatID, &payload.Read)
}
