package outbox

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Relay struct {
	pool      *pgxpool.Pool
	producer  sarama.SyncProducer
	logger    *slog.Logger
	interval  time.Duration
	batchSize int32
}

func NewRelay(pool *pgxpool.Pool, producer sarama.SyncProducer, logger *slog.Logger, interval time.Duration, batchSize int32) *Relay {
	return &Relay{
		pool:      pool,
		producer:  producer,
		logger:    logger,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				n, err := r.processBatch(ctx)
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						r.logger.Error("outbox relay batch failed", "error", err.Error())
					}
					break
				}
				if n == 0 {
					break
				}
			}
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	qtx := queries.New(tx)

	rows, err := qtx.FetchUnpublishedOutboxMessages(ctx, r.batchSize)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	published := make([]uuid.UUID, 0, len(rows))
	var publishErr error
	for _, row := range rows {
		message := &sarama.ProducerMessage{
			Topic: row.Topic,
			Value: sarama.ByteEncoder(row.Payload),
		}
		if row.AggregateID.Valid {
			key := row.AggregateID.Bytes
			message.Key = sarama.ByteEncoder(key[:])
		}

		if _, _, err := r.producer.SendMessage(message); err != nil {
			publishErr = err
			break
		}

		published = append(published, row.ID)
	}

	if len(published) > 0 {
		if err := qtx.MarkOutboxMessagesPublished(ctx, published); err != nil {
			return 0, err
		}
		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}
	}

	if publishErr != nil {
		return len(published), publishErr
	}

	return len(published), nil
}

func (r *Relay) Close() error {
	return r.producer.Close()
}
