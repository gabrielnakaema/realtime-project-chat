package notification

import (
	"context"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/publisher"
)

type KafkaNotifier struct {
	publisher *publisher.Publisher
}

func NewKafkaNotifier(publisher *publisher.Publisher) *KafkaNotifier {
	return &KafkaNotifier{publisher: publisher}
}

func (kn *KafkaNotifier) SendNotification(ctx context.Context, notification *domain.Notification) error {
	payload := events.NotificationCreatedPayload{Notification: *notification}
	return kn.publisher.Publish(ctx, events.NotificationCreated, &payload)
}
