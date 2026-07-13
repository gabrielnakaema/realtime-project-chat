package subscriber

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/google/uuid"
)

type RealtimeNotifier interface {
	SendMessages(context.Context, *domain.ChatMessage) error
	SendReadUpdate(context.Context, uuid.UUID, *domain.ChatMessageRead) error
	SendChatProjectMemberCreated(context.Context, *domain.ProjectMember, uuid.UUID) error
	SendChatProjectMemberRemoved(context.Context, *domain.ProjectMember, uuid.UUID) error
}

// RealtimeSubscriber owns chat-related fan-out for the standalone websocket
// process. Its consumer group is intentionally separate from chat.subscriber:
// the API consumer persists chat state, while this consumer only delivers it.
type RealtimeSubscriber struct {
	logger     *slog.Logger
	subscriber *Subscriber
	notifier   RealtimeNotifier
}

func NewRealtimeSubscriber(ctx context.Context, cfg *config.Config, logger *slog.Logger, notifier RealtimeNotifier) (*RealtimeSubscriber, error) {
	sub, err := NewSubscriber(cfg, "websocket.realtime.subscriber")
	if err != nil {
		return nil, domain.ServerError("failed to create realtime subscriber", err)
	}

	realtimeSubscriber := &RealtimeSubscriber{
		logger:     logger,
		subscriber: sub,
		notifier:   notifier,
	}

	topics := []events.Topic{
		events.ChatMessageCreated,
		events.ChatMessageRead,
		events.ChatMemberCreated,
		events.ChatMemberRemoved,
	}
	if err := sub.Subscribe(ctx, topics, realtimeSubscriber.handleEvents, logger); err != nil {
		_ = sub.Close()
		return nil, domain.ServerError("failed to subscribe to realtime events", err)
	}

	return realtimeSubscriber, nil
}

func (s *RealtimeSubscriber) Close() error {
	return s.subscriber.Close()
}

func (s *RealtimeSubscriber) handleEvents(ctx context.Context, message Message) error {
	switch message.Topic {
	case events.ChatMessageCreated:
		var payload events.ChatMessageCreatedPayload
		if err := json.Unmarshal(message.Value, &payload); err != nil {
			return domain.ServerError("failed to unmarshal chat message", err)
		}
		return s.notifier.SendMessages(ctx, &payload.ChatMessage)
	case events.ChatMessageRead:
		var payload events.ChatMessageReadPayload
		if err := json.Unmarshal(message.Value, &payload); err != nil {
			return domain.ServerError("failed to unmarshal chat message read", err)
		}
		return s.notifier.SendReadUpdate(ctx, payload.ChatID, &payload.Read)
	case events.ChatMemberCreated:
		var payload events.ChatMemberCreatedPayload
		if err := json.Unmarshal(message.Value, &payload); err != nil {
			return domain.ServerError("failed to unmarshal chat member created", err)
		}
		return s.notifier.SendChatProjectMemberCreated(ctx, &payload.ProjectMember, payload.ChatMember.ChatId)
	case events.ChatMemberRemoved:
		var payload events.ChatMemberRemovedPayload
		if err := json.Unmarshal(message.Value, &payload); err != nil {
			return domain.ServerError("failed to unmarshal chat member removed", err)
		}
		return s.notifier.SendChatProjectMemberRemoved(ctx, &payload.ProjectMember, payload.ChatMember.ChatId)
	default:
		return nil
	}
}
