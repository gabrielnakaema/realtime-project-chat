package repository

import (
	"context"
	"errors"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (ur *UserRepository) Create(ctx context.Context, user *domain.User) error {
	q := queries.New(ur.pool)

	params := queries.CreateUserParams{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}

	id, err := q.CreateUser(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				if pgErr.ConstraintName == "users_email_key" {
					return domain.DuplicateEntryError("user email is already taken")
				}
				return err
			default:
				return err
			}
		}

		return err
	}

	user.Id = id

	return nil
}

func (ur *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := queries.New(ur.pool)

	userResult, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("user not found")
		}
		return nil, err
	}

	user := domain.User{
		Id:        userResult.ID,
		Email:     userResult.Email,
		Name:      userResult.Name,
		Password:  userResult.Password,
		CreatedAt: userResult.CreatedAt.Time,
	}

	return &user, nil
}

func (ur *UserRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	q := queries.New(ur.pool)

	userResult, err := q.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("user not found")
		}
		return nil, err
	}

	user := domain.User{
		Id:        userResult.ID,
		Email:     userResult.Email,
		Name:      userResult.Name,
		Password:  userResult.Password,
		CreatedAt: userResult.CreatedAt.Time,
	}

	return &user, nil
}

func (ur *UserRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	q := queries.New(ur.pool)

	tokenResult, err := q.GetRefreshTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("refresh token not found")
		}
		return nil, err
	}

	refreshToken := domain.RefreshToken{
		Id:        tokenResult.ID,
		UserId:    tokenResult.UserID,
		Token:     tokenResult.Token,
		Active:    tokenResult.Active,
		CreatedAt: tokenResult.CreatedAt.Time,
		ExpiresAt: tokenResult.ExpiresAt.Time,
	}

	return &refreshToken, nil
}

func (ur *UserRepository) CreateRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error {
	q := queries.New(ur.pool)

	params := queries.CreateRefreshTokenParams{
		UserID: refreshToken.UserId,
		ExpiresAt: pgtype.Timestamptz{
			Time:  refreshToken.ExpiresAt,
			Valid: true,
		},
		Token:  refreshToken.Token,
		Active: refreshToken.Active,
	}

	tokenId, err := q.CreateRefreshToken(ctx, params)
	if err != nil {
		return err
	}

	refreshToken.Id = tokenId

	return nil
}

func (ur *UserRepository) UpdateRefreshTokenActive(ctx context.Context, refreshToken *domain.RefreshToken) error {
	q := queries.New(ur.pool)

	params := queries.UpdateRefreshTokenParams{
		Token:  refreshToken.Token,
		Active: refreshToken.Active,
	}

	err := q.UpdateRefreshToken(ctx, params)
	return err
}
