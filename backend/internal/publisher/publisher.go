package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
)

type Publisher struct {
	config   *config.Config
	producer sarama.AsyncProducer
}

func NewPublisher(config *config.Config) (*Publisher, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5

	producer, err := sarama.NewAsyncProducer(config.PubsubBrokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		config:   config,
		producer: producer,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, topic events.Topic, payload interface{}) error {
	if !topic.Valid() {
		// Server error since it's a programmer error
		return domain.ServerError("invalid topic", fmt.Errorf("invalid topic provided to Publish function: %s", topic.String()))
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return domain.ServerError("failed to marshal payload", err)
	}

	message := &sarama.ProducerMessage{
		Topic: topic.String(),
		Value: sarama.ByteEncoder(bytes),
	}

	p.producer.Input() <- message

	select {
	case <-p.producer.Successes():
		return nil
	case err = <-p.producer.Errors():
		return err
	case <-ctx.Done():
		return errors.New("context done")
	}
}

func (p *Publisher) Close() error {
	return p.producer.Close()
}
