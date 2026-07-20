package user

import (
	"context"
	"errors"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	userqueries "github.com/gabrielnakaema/project-chat/internal/user/queries"
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
	q := userqueries.New(ur.pool)

	params := userqueries.CreateUserParams{
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
					return apperr.DuplicateEntryError("user email is already taken")
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
	q := userqueries.New(ur.pool)

	userResult, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFoundError("user not found")
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
	q := userqueries.New(ur.pool)

	userResult, err := q.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFoundError("user not found")
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

func (ur *UserRepository) UpdatePassword(ctx context.Context, user *domain.User) error {
	q := userqueries.New(ur.pool)

	affected, err := q.UpdateUserPassword(ctx, userqueries.UpdateUserPasswordParams{
		ID:       user.Id,
		Password: user.Password,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return apperr.NotFoundError("user not found")
	}

	return nil
}

func (ur *UserRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	q := userqueries.New(ur.pool)

	tokenResult, err := q.GetRefreshTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFoundError("refresh token not found")
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
	q := userqueries.New(ur.pool)

	params := userqueries.CreateRefreshTokenParams{
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

func (ur *UserRepository) ListUsers(ctx context.Context, excludeId uuid.UUID) ([]domain.User, error) {
	q := userqueries.New(ur.pool)

	rows, err := q.ListUsers(ctx, excludeId)
	if err != nil {
		return nil, err
	}

	users := []domain.User{}
	for _, row := range rows {
		users = append(users, domain.User{
			Id:        row.ID,
			Name:      row.Name,
			Email:     row.Email,
			CreatedAt: row.CreatedAt.Time,
		})
	}

	return users, nil
}

func (ur *UserRepository) UpdateRefreshTokenActive(ctx context.Context, refreshToken *domain.RefreshToken) error {
	q := userqueries.New(ur.pool)

	params := userqueries.UpdateRefreshTokenParams{
		Token:  refreshToken.Token,
		Active: refreshToken.Active,
	}

	err := q.UpdateRefreshToken(ctx, params)
	return err
}
