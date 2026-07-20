package kafka

import (
	"github.com/IBM/sarama"
	"github.com/gabrielnakaema/project-chat/internal/platform/config"
)

func NewSyncPublisher(cfg *config.Config) (sarama.SyncProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5

	return sarama.NewSyncProducer(cfg.PubsubBrokers, saramaConfig)
}
