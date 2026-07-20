package messaging

import (
	"context"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/platform/config"
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
			select {
			case <-ctx.Done():
				logger.Info("subscriber context cancelled, stopping consume loop", "topics", topics)
				return
			default:
			}

			err := s.consumer.Consume(ctx, topics, &consumerGroupHandler{handler: handler, logger: logger})
			if err != nil {
				logger.Error("error consuming topic", "error", err.Error())
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
			}
		}
	}()

	return nil
}

func (s *Subscriber) Close() error {
	return s.consumer.Close()
}

const maxHandlerRetries = 3
const handlerRetryBackoff = 200 * time.Millisecond

type consumerGroupHandler struct {
	handler MessageHandler
	logger  *slog.Logger
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

		var err error
		for attempt := 1; attempt <= maxHandlerRetries; attempt++ {
			err = h.handler(session.Context(), m)
			if err == nil {
				break
			}

			if attempt < maxHandlerRetries {
				select {
				case <-session.Context().Done():
					return nil
				case <-time.After(handlerRetryBackoff):
				}
			}
		}

		if err != nil {
			h.logger.Error(
				"dropping message after exhausting handler retries",
				"topic", message.Topic,
				"partition", message.Partition,
				"offset", message.Offset,
				"attempts", maxHandlerRetries,
				"error", err.Error(),
			)
		}

		session.MarkMessage(message, "")
	}

	return nil
}
