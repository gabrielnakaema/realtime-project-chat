package subscriber

import (
	"context"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/events"
)

type Subscriber struct {
	consumer sarama.ConsumerGroup
}

func NewSubscriber(config *config.Config, groupId string) (*Subscriber, error) {
	saramaConfig := sarama.NewConfig()

	consumer, err := sarama.NewConsumerGroup(config.PubsubBrokers, groupId, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		consumer: consumer,
	}, nil
}

type Message struct {
	Topic     events.Topic
	Key       []byte
	Value     []byte
	Timestamp time.Time
	Metadata  map[string]string
}

type MessageHandler func(ctx context.Context, message Message) error

func (s *Subscriber) Subscribe(ctx context.Context, topic []events.Topic, handler MessageHandler, logger *slog.Logger) error {
	topics := []string{}
	for _, t := range topic {
		topics = append(topics, t.String())
	}

	go func() {
		for {
			err := s.consumer.Consume(ctx, topics, &consumerGroupHandler{handler: handler})
			if err != nil {
				logger.Error("error consuming topic", "error", err.Error())
			}
		}
	}()

	return nil
}

type consumerGroupHandler struct {
	handler MessageHandler
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		m := Message{
			Topic:     events.Topic(message.Topic),
			Key:       message.Key,
			Value:     message.Value,
			Timestamp: message.Timestamp,
		}

		err := h.handler(session.Context(), m)
		if err != nil {
			return err
		}

		session.MarkMessage(message, "")
	}

	return nil
}
