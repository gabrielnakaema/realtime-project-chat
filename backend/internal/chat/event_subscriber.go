package chat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/config"
	"github.com/gabrielnakaema/project-chat/internal/platform/messaging"
)

type EventSubscriber struct {
	logger      *slog.Logger
	subscriber  *messaging.Subscriber
	chatService *Service
}

func NewEventSubscriber(ctx context.Context, cfg *config.Config, logger *slog.Logger, chatService *Service) (*EventSubscriber, error) {
	sub, err := messaging.NewSubscriber(cfg, "chat.subscriber")
	if err != nil {
		return nil, apperr.ServerError("failed to create chat subscriber", err)
	}

	eventSubscriber := &EventSubscriber{
		subscriber:  sub,
		logger:      logger,
		chatService: chatService,
	}

	topics := []events.Topic{events.ProjectCreated, events.ProjectMemberCreated, events.ProjectMemberRemoved, events.ChatMemberCreated, events.ChatMemberViewed}

	err = sub.Subscribe(ctx, topics, eventSubscriber.handleChatEvents, eventSubscriber.logger)
	if err != nil {
		return nil, apperr.ServerError("failed to subscribe to chat events", err)
	}

	return eventSubscriber, nil
}

func (cs *EventSubscriber) Close() error {
	return cs.subscriber.Close()
}

func (cs *EventSubscriber) handleChatEvents(ctx context.Context, message messaging.Message) error {
	switch message.Topic {
	case events.ProjectCreated:
		return cs.handleProjectCreated(ctx, message)
	case events.ProjectMemberCreated:
		return cs.handleProjectMemberCreated(ctx, message)
	case events.ProjectMemberRemoved:
		return cs.handleProjectMemberRemoved(ctx, message)
	case events.ChatMemberCreated:
		return cs.handleChatMemberCreated(ctx, message)
	case events.ChatMemberViewed:
		return cs.handleChatMemberViewed(ctx, message)
	}

	return nil
}

func (cs *EventSubscriber) handleChatMemberViewed(ctx context.Context, message messaging.Message) error {
	var payload events.ChatMemberViewedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return apperr.ServerError("failed to unmarshal chat member", err)
	}

	err = cs.chatService.UpdateMemberLastSeenAt(ctx, payload.ChatMember.UserId, payload.ChatMember.ChatId)
	if err != nil {
		return err
	}

	return nil
}

func (cs *EventSubscriber) handleProjectCreated(ctx context.Context, message messaging.Message) error {
	var payload events.ProjectCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return apperr.ServerError("failed to unmarshal project", err)
	}

	err = cs.chatService.CreateChatFromProject(ctx, &payload.Project)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.ServerErrorCode && domainErr.Cause != nil {
				cs.logger.Error("failed to create chat from project", "error", domainErr.Cause.Error())
				return nil
			}
			return err
		}
		cs.logger.Error("failed to create chat from project", "error", err.Error())
		return apperr.ServerError("failed to create chat from project", err)
	}

	return nil
}

func (cs *EventSubscriber) handleProjectMemberCreated(ctx context.Context, message messaging.Message) error {
	var payload events.ProjectMemberCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return apperr.ServerError("failed to unmarshal project member", err)
	}

	err = cs.chatService.CreateMemberFromProjectMember(ctx, &payload.ProjectMember)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				cs.logger.Info("chat not found, skipping creation of chat member from project member", "project_member", payload.ProjectMember)
				return nil
			}
			return err
		}

		cs.logger.Error("failed to create member from project member", "error", err)
		return apperr.ServerError("failed to create member from project member", err)
	}

	return nil
}

func (cs *EventSubscriber) handleProjectMemberRemoved(ctx context.Context, message messaging.Message) error {
	var payload events.ProjectMemberRemovedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return apperr.ServerError("failed to unmarshal project member removed payload", err)
	}

	_, err := cs.chatService.RemoveMemberFromProjectMember(ctx, &payload.ProjectMember)
	if err != nil {
		return err
	}

	return nil
}

func (cs *EventSubscriber) handleChatMemberCreated(ctx context.Context, message messaging.Message) error {
	var payload events.ChatMemberCreatedPayload
	err := json.Unmarshal(message.Value, &payload)
	if err != nil {
		return apperr.ServerError("failed to unmarshal chat member", err)
	}

	err = cs.chatService.CreateJoinedMessage(ctx, &payload.ChatMember)
	if err != nil {
		return apperr.ServerError("failed to update member last seen at", err)
	}

	return nil
}
