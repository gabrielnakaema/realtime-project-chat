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
)

type MessageNotifier interface {
	SendMessages(ctx context.Context, message *domain.ChatMessage) error
}

type ChatSubscriber struct {
	logger      *slog.Logger
	subscriber  *Subscriber
	chatService *service.ChatService
	notifier    MessageNotifier
}

func NewChatSubscriber(config *config.Config, logger *slog.Logger, chatService *service.ChatService, notifier MessageNotifier) (*ChatSubscriber, error) {
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

	topics := []events.Topic{events.ProjectCreated, events.ProjectMemberCreated, events.ChatMemberCreated, events.ChatMessageCreated}

	err = subscriber.Subscribe(context.Background(), topics, chatSubscriber.handleChatEvents, chatSubscriber.logger)
	if err != nil {
		return nil, domain.ServerError("failed to subscribe to chat events", err)
	}

	return chatSubscriber, nil
}

func (cs *ChatSubscriber) handleChatEvents(ctx context.Context, message Message) error {
	switch message.Topic {
	case events.ProjectCreated:
		return cs.handleProjectCreated(ctx, message)
	case events.ProjectMemberCreated:
		return cs.handleProjectMemberCreated(ctx, message)
	case events.ChatMemberCreated:
		return cs.handleChatMemberCreated(ctx, message)
	case events.ChatMessageCreated:
		return cs.handleChatMessageCreated(ctx, message)
	}

	return nil
}

func (cs *ChatSubscriber) handleProjectCreated(ctx context.Context, message Message) error {
	var project domain.Project
	err := json.Unmarshal(message.Value, &project)
	if err != nil {
		return domain.ServerError("failed to unmarshal project", err)
	}

	err = cs.chatService.CreateChatFromProject(ctx, &project)
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
	var projectMember domain.ProjectMember
	err := json.Unmarshal(message.Value, &projectMember)
	if err != nil {
		return domain.ServerError("failed to unmarshal project member", err)
	}

	err = cs.chatService.CreateMemberFromProjectMember(ctx, &projectMember)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				cs.logger.Info("chat not found, skipping creation of chat member from project member", "project_member", projectMember)
				return nil
			}
			return err
		}

		cs.logger.Error("failed to create member from project member", "error", err)
		return domain.ServerError("failed to create member from project member", err)
	}

	return nil
}

func (cs *ChatSubscriber) handleChatMemberCreated(ctx context.Context, message Message) error {
	var chatMember domain.ChatMember
	err := json.Unmarshal(message.Value, &chatMember)
	if err != nil {
		return domain.ServerError("failed to unmarshal chat member", err)
	}

	err = cs.chatService.CreateJoinedMessage(ctx, &chatMember)
	if err != nil {
		return domain.ServerError("failed to update member last seen at", err)
	}

	return nil
}

func (cs *ChatSubscriber) handleChatMessageCreated(ctx context.Context, message Message) error {
	var chatMessage domain.ChatMessage
	err := json.Unmarshal(message.Value, &chatMessage)
	if err != nil {
		return domain.ServerError("failed to unmarshal chat message", err)
	}

	err = cs.notifier.SendMessages(ctx, &chatMessage)
	if err != nil {
		return domain.ServerError("failed to send messages", err)
	}

	return nil
}
