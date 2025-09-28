package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/events"
)

type Publisher struct {
	config   *config.Config
	producer sarama.AsyncProducer
	logger   *slog.Logger
	wg       sync.WaitGroup
	done     chan struct{}
}

func NewPublisher(config *config.Config, logger *slog.Logger) (*Publisher, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5

	producer, err := sarama.NewAsyncProducer(config.PubsubBrokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	publisher := &Publisher{
		config:   config,
		producer: producer,
		logger:   logger,
		wg:       sync.WaitGroup{},
		done:     make(chan struct{}),
	}

	publisher.wg.Add(2)
	go publisher.handleSuccesses()
	go publisher.handleErrors()

	return publisher, nil
}

func (p *Publisher) handleSuccesses() {
	defer p.wg.Done()
	for {
		select {
		case success := <-p.producer.Successes():
			p.logger.Debug("message sent successfully", "topic", success.Topic, "partition", success.Partition, "offset", success.Offset)
		case <-p.done:
			return
		}
	}
}

func (p *Publisher) handleErrors() {
	defer p.wg.Done()
	for {
		select {
		case err := <-p.producer.Errors():
			p.logger.Error("producer error", "topic", err.Msg.Topic, "error", err.Err.Error())
		case <-p.done:
			return
		}
	}
}

func (p *Publisher) Publish(ctx context.Context, topic events.Topic, payload interface{}) error {
	if !topic.Valid() {
		return errors.New("invalid topic provided")
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return errors.New("failed to marshal payload")
	}

	message := &sarama.ProducerMessage{
		Topic: topic.String(),
		Value: sarama.ByteEncoder(bytes),
	}

	select {
	case p.producer.Input() <- message:
		return nil
	case <-ctx.Done():
		return errors.New("context done")
	}
}

func (p *Publisher) Close() error {
	close(p.done)
	p.wg.Wait()
	return p.producer.Close()
}
