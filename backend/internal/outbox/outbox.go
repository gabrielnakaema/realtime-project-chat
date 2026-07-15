package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Message struct {
	Topic       events.Topic
	AggregateID uuid.UUID
	Payload     any
}

func Enqueue(ctx context.Context, dbtx queries.DBTX, msgs ...Message) error {
	q := queries.New(dbtx)
	for _, m := range msgs {
		if !m.Topic.Valid() {
			return fmt.Errorf("outbox: invalid topic %q", m.Topic)
		}

		payload, err := json.Marshal(m.Payload)
		if err != nil {
			return fmt.Errorf("outbox: marshal payload for %q: %w", m.Topic, err)
		}

		if err := q.InsertOutboxMessage(ctx, queries.InsertOutboxMessageParams{
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
