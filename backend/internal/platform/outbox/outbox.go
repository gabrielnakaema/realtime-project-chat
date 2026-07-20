package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	outboxqueries "github.com/gabrielnakaema/project-chat/internal/platform/outbox/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Topic interface {
	String() string
	Valid() bool
}

type Message struct {
	Topic       Topic
	AggregateID uuid.UUID
	Payload     any
}

func Enqueue(ctx context.Context, dbtx outboxqueries.DBTX, msgs ...Message) error {
	q := outboxqueries.New(dbtx)
	for _, m := range msgs {
		if m.Topic == nil || !m.Topic.Valid() {
			return fmt.Errorf("outbox: invalid topic %q", m.Topic)
		}

		payload, err := json.Marshal(m.Payload)
		if err != nil {
			return fmt.Errorf("outbox: marshal payload for %q: %w", m.Topic, err)
		}

		if err := q.InsertOutboxMessage(ctx, outboxqueries.InsertOutboxMessageParams{
			Topic:       m.Topic.String(),
			AggregateID: optionalUUID(m.AggregateID),
			Payload:     payload,
		}); err != nil {
			return fmt.Errorf("outbox: insert message for %q: %w", m.Topic, err)
		}
	}

	return nil
}

func optionalUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{}
	}

	return pgtype.UUID{Bytes: id, Valid: true}
}
