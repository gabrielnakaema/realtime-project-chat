package notification

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/subscriber"
)

type ForwardSubscriber struct {
	logger     *slog.Logger
	subscriber *subscriber.Subscriber
	notifier   Notifier
}

func NewForwardSubscriber(ctx context.Context, cfg *config.Config, logger *slog.Logger, notifier Notifier) (*ForwardSubscriber, error) {
	sub, err := subscriber.NewSubscriber(cfg, "notification-forwarder.subscriber")
	if err != nil {
		return nil, err
	}

	forwardSubscriber := &ForwardSubscriber{
		logger:     logger,
		subscriber: sub,
		notifier:   notifier,
	}

	topics := []events.Topic{events.NotificationCreated}
	if err := sub.Subscribe(ctx, topics, forwardSubscriber.handleNotificationCreated, forwardSubscriber.logger); err != nil {
		return nil, err
	}

	return forwardSubscriber, nil
}

func (fs *ForwardSubscriber) Close() error {
	return fs.subscriber.Close()
}

func (fs *ForwardSubscriber) handleNotificationCreated(ctx context.Context, message subscriber.Message) error {
	var payload events.NotificationCreatedPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ServerError("failed to unmarshal notification created payload", err)
	}

	return fs.notifier.SendNotification(ctx, &payload.Notification)
}
