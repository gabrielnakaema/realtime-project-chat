package apikey

import (
	"context"
	"errors"
	"fmt"
	"time"

	apikeyqueries "github.com/gabrielnakaema/project-chat/internal/apikey/queries"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MCPAPIKeyRepository struct {
	pool *pgxpool.Pool
}

func NewMCPAPIKeyRepository(pool *pgxpool.Pool) *MCPAPIKeyRepository {
	return &MCPAPIKeyRepository{pool: pool}
}

func (r *MCPAPIKeyRepository) Create(ctx context.Context, key *domain.MCPAPIKey) error {
	q := apikeyqueries.New(r.pool)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	key.ID, err = qtx.CreateMCPAPIKey(ctx, apikeyqueries.CreateMCPAPIKeyParams{
		UserID:     key.UserID,
		Name:       key.Name,
		KeyPrefix:  key.KeyPrefix,
		SecretHash: key.SecretHash,
		CreatedAt: pgtype.Timestamptz{
			Time:  key.CreatedAt,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	for _, scope := range key.Scopes {
		if err := qtx.CreateMCPAPIKeyScope(ctx, apikeyqueries.CreateMCPAPIKeyScopeParams{
			ApiKeyID: key.ID,
			Scope:    string(scope),
			CreatedAt: pgtype.Timestamptz{
				Time:  key.CreatedAt,
				Valid: true,
			},
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *MCPAPIKeyRepository) Update(ctx context.Context, key *domain.MCPAPIKey) error {
	q := apikeyqueries.New(r.pool)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	affected, err := qtx.UpdateMCPAPIKeyName(ctx, apikeyqueries.UpdateMCPAPIKeyNameParams{
		ID:     key.ID,
		UserID: key.UserID,
		Name:   key.Name,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return apperr.NotFoundError("mcp api key not found")
	}

	if err := qtx.DeleteMCPAPIKeyScopes(ctx, key.ID); err != nil {
		return err
	}

	now := time.Now()
	createdAt := pgtype.Timestamptz{
		Time:  now,
		Valid: true,
	}
	for _, scope := range key.Scopes {
		if err := qtx.CreateMCPAPIKeyScope(ctx, apikeyqueries.CreateMCPAPIKeyScopeParams{
			ApiKeyID:  key.ID,
			Scope:     string(scope),
			CreatedAt: createdAt,
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *MCPAPIKeyRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.MCPAPIKey, error) {
	q := apikeyqueries.New(r.pool)

	rows, err := q.ListMCPAPIKeysByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	keys := make([]domain.MCPAPIKey, 0, len(rows))
	for _, row := range rows {
		key, err := mapMCPAPIKeyRow(
			row.ID,
			row.UserID,
			row.Name,
			row.KeyPrefix,
			row.SecretHash,
			row.CreatedAt,
			row.LastUsedAt,
			row.RevokedAt,
			row.Scopes,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (r *MCPAPIKeyRepository) GetByIDForUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.MCPAPIKey, error) {
	q := apikeyqueries.New(r.pool)

	row, err := q.GetMCPAPIKeyByIDForUser(ctx, apikeyqueries.GetMCPAPIKeyByIDForUserParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFoundError("mcp api key not found")
		}
		return nil, err
	}

	key, err := mapMCPAPIKeyRow(
		row.ID,
		row.UserID,
		row.Name,
		row.KeyPrefix,
		row.SecretHash,
		row.CreatedAt,
		row.LastUsedAt,
		row.RevokedAt,
		row.Scopes,
	)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (r *MCPAPIKeyRepository) GetByPrefix(ctx context.Context, prefix string) (*domain.MCPAPIKey, error) {
	q := apikeyqueries.New(r.pool)

	row, err := q.GetMCPAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFoundError("mcp api key not found")
		}
		return nil, err
	}

	key, err := mapMCPAPIKeyRow(
		row.ID,
		row.UserID,
		row.Name,
		row.KeyPrefix,
		row.SecretHash,
		row.CreatedAt,
		row.LastUsedAt,
		row.RevokedAt,
		row.Scopes,
	)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (r *MCPAPIKeyRepository) Revoke(ctx context.Context, id uuid.UUID, userID uuid.UUID, revokedAt time.Time) error {
	q := apikeyqueries.New(r.pool)

	affected, err := q.RevokeMCPAPIKey(ctx, apikeyqueries.RevokeMCPAPIKeyParams{
		ID:     id,
		UserID: userID,
		RevokedAt: pgtype.Timestamptz{
			Time:  revokedAt,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return apperr.NotFoundError("mcp api key not found")
	}

	return nil
}

func (r *MCPAPIKeyRepository) TouchLastUsedAt(ctx context.Context, id uuid.UUID, now time.Time, minInterval time.Duration) error {
	q := apikeyqueries.New(r.pool)
	cutoff := now.Add(-minInterval)

	return q.TouchMCPAPIKeyLastUsedAt(ctx, apikeyqueries.TouchMCPAPIKeyLastUsedAtParams{
		ID: id,
		LastUsedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		CutoffAt: pgtype.Timestamptz{
			Time:  cutoff,
			Valid: true,
		},
	})
}

func mapMCPAPIKeyRow(
	id uuid.UUID,
	userID uuid.UUID,
	name string,
	keyPrefix string,
	secretHash string,
	createdAt pgtype.Timestamptz,
	lastUsedAt pgtype.Timestamptz,
	revokedAt pgtype.Timestamptz,
	rawScopes any,
) (domain.MCPAPIKey, error) {
	scopes, err := normalizeMCPAPIKeyScopes(rawScopes)
	if err != nil {
		return domain.MCPAPIKey{}, err
	}

	key := domain.MCPAPIKey{
		ID:         id,
		UserID:     userID,
		Name:       name,
		KeyPrefix:  keyPrefix,
		SecretHash: secretHash,
		Scopes:     make([]domain.MCPAPIScope, 0, len(scopes)),
		CreatedAt:  createdAt.Time,
	}

	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}

	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}

	for _, scope := range scopes {
		key.Scopes = append(key.Scopes, domain.MCPAPIScope(scope))
	}

	return key, nil
}

func normalizeMCPAPIKeyScopes(rawScopes any) ([]string, error) {
	switch scopes := rawScopes.(type) {
	case nil:
		return nil, nil
	case []string:
		return scopes, nil
	case []any:
		values := make([]string, 0, len(scopes))
		for _, scope := range scopes {
			value, ok := scope.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected mcp api key scope value type %T", scope)
			}
			values = append(values, value)
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unexpected mcp api key scopes type %T", rawScopes)
	}
}
