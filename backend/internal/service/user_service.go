package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	UpdateRefreshTokenActive(ctx context.Context, refreshToken *domain.RefreshToken) error
	ListUsers(ctx context.Context, excludeId uuid.UUID) ([]domain.User, error)
}

type jwtProvider interface {
	Generate(sub string, exp time.Time, claims map[string]string) (string, error)
}

type UserService struct {
	jwtProvider    jwtProvider
	userRepository userRepository
}

func NewUserService(jwtProvider jwtProvider, userRepository userRepository) *UserService {
	return &UserService{
		jwtProvider:    jwtProvider,
		userRepository: userRepository,
	}
}

type CreateUserRequest struct {
	Name     string
	Email    string
	Password string
}

func (us *UserService) Create(ctx context.Context, request CreateUserRequest) (*domain.User, error) {
	hashed, err := HashPassword(request.Password)
	if err != nil {
		return nil, domain.ServerError("failed to hash password", err)
	}

	user := domain.User{
		Name:      request.Name,
		Email:     request.Email,
		Password:  hashed,
		CreatedAt: time.Now(),
	}

	err = us.userRepository.Create(ctx, &user)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.DuplicateEntryErrorCode {
				return nil, domain.DuplicateEntryError("user email is already taken")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to create user", err)
	}

	return &user, nil
}

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

func (us *UserService) Login(ctx context.Context, request LoginRequest) (*LoginResult, error) {
	const INVALID_CREDENTIALS_ERROR_MESSAGE = "invalid credentials"

	user, err := us.userRepository.GetByEmail(ctx, request.Email)
	if err != nil {
		return nil, domain.UnauthorizedError(INVALID_CREDENTIALS_ERROR_MESSAGE)
	}

	equal, err := CompareHash(request.Password, user.Password)
	if err != nil {
		return nil, domain.UnauthorizedError(INVALID_CREDENTIALS_ERROR_MESSAGE)
	}

	if !equal {
		return nil, domain.UnauthorizedError(INVALID_CREDENTIALS_ERROR_MESSAGE)
	}

	claims := make(map[string]string)

	now := time.Now()

	token, err := us.jwtProvider.Generate(user.Id.String(), now.Add(domain.AccessTokenTTL), claims)
	if err != nil {
		return nil, domain.ServerError("error while generating token", err)
	}

	refreshTokenToken, err := GenerateRefreshToken(domain.RefreshTokenByteLength)
	if err != nil {
		return nil, domain.ServerError("error while generating refresh token", err)
	}

	refreshToken := domain.NewRefreshToken(user.Id, refreshTokenToken, now)

	err = us.userRepository.CreateRefreshToken(ctx, &refreshToken)
	if err != nil {
		return nil, domain.ServerError("error while saving refresh token", err)
	}

	result := LoginResult{
		AccessToken:  token,
		RefreshToken: refreshToken.Token,
		User:         user,
	}

	return &result, nil
}

func (us *UserService) GetMe(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return us.userRepository.GetById(ctx, id)
}

func (us *UserService) ListUsers(ctx context.Context, excludeId uuid.UUID) ([]domain.User, error) {
	return us.userRepository.ListUsers(ctx, excludeId)
}

type RefreshTokenRequest struct {
	Token string
}

func (us *UserService) RefreshToken(ctx context.Context, request RefreshTokenRequest) (*LoginResult, error) {
	now := time.Now()

	refreshToken, err := us.userRepository.GetRefreshToken(ctx, request.Token)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.UnauthorizedError("invalid refresh token")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("error while getting refresh token", err)
	}

	if !refreshToken.IsUsableAt(now) {
		return nil, domain.UnauthorizedError("invalid refresh token")
	}

	user, err := us.userRepository.GetById(ctx, refreshToken.UserId)
	if err != nil {
		return nil, domain.UnauthorizedError("invalid refresh token")
	}

	accessToken, err := us.jwtProvider.Generate(user.Id.String(), now.Add(domain.AccessTokenTTL), make(map[string]string))
	if err != nil {
		return nil, domain.ServerError("error while generating token", err)
	}

	newRefreshTokenToken, err := GenerateRefreshToken(domain.RefreshTokenByteLength)
	if err != nil {
		return nil, domain.ServerError("error while generating refresh token", err)
	}

	newRefreshToken := domain.NewRefreshToken(user.Id, newRefreshTokenToken, now)

	err = us.userRepository.CreateRefreshToken(ctx, &newRefreshToken)
	if err != nil {
		return nil, domain.ServerError("error while creating new refresh token", err)
	}

	refreshToken.Deactivate()

	_ = us.userRepository.UpdateRefreshTokenActive(ctx, refreshToken)

	result := LoginResult{
		User:         user,
		RefreshToken: newRefreshTokenToken,
		AccessToken:  accessToken,
	}

	return &result, nil
}

func (us *UserService) Logout(ctx context.Context, userId uuid.UUID, token string) error {
	refreshToken, err := us.userRepository.GetRefreshToken(ctx, token)
	if err != nil {
		return err
	}

	if refreshToken.UserId != userId {
		return domain.ForbiddenError("invalid refresh token")
	}

	refreshToken.Deactivate()

	return us.userRepository.UpdateRefreshTokenActive(ctx, refreshToken)
}

func GenerateRefreshToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

func HashPassword(plaintext string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), 10)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func CompareHash(plaintext string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	if err != nil {
		return false, err
	}

	return true, nil
}
