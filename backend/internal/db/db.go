package db

import (
	"context"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(config *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
