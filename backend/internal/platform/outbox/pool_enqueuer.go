package outbox

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolEnqueuer struct {
	pool *pgxpool.Pool
}

func NewPoolEnqueuer(pool *pgxpool.Pool) *PoolEnqueuer {
	return &PoolEnqueuer{pool: pool}
}

func (e *PoolEnqueuer) Enqueue(ctx context.Context, msgs ...Message) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := Enqueue(ctx, tx, msgs...); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
