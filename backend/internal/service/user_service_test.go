package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) CreateRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *mockUserRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *mockUserRepository) UpdateRefreshTokenActive(ctx context.Context, refreshToken *domain.RefreshToken) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

type mockJWTProvider struct {
	mock.Mock
}

func (m *mockJWTProvider) Generate(sub string, exp time.Time, claims map[string]string) (string, error) {
	args := m.Called(sub, exp, claims)
	return args.String(0), args.Error(1)
}

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name          string
		request       service.CreateUserRequest
		mockSetup     func(*mockUserRepository)
		expectedUser  *domain.User
		expectedError error
		shouldSucceed bool
	}{
		{
			name: "successful user creation",
			request: service.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(repo *mockUserRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
					user := args.Get(1).(*domain.User)
					user.Id = uuid.New()
				})
			},
			shouldSucceed: true,
		},
		{
			name: "duplicate email error",
			request: service.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(repo *mockUserRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(domain.DuplicateEntryError("user email is already taken"))
			},
			expectedError: domain.DuplicateEntryError("user email is already taken"),
			shouldSucceed: false,
		},
		{
			name: "repository error",
			request: service.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(repo *mockUserRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))
			},
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			mockJWT := &mockJWTProvider{}
			tt.mockSetup(mockRepo)

			service := service.NewUserService(mockJWT, mockRepo)
			ctx := context.Background()

			user, err := service.Create(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Name, user.Name)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.NotEmpty(t, user.Password)
				assert.NotEqual(t, tt.request.Password, user.Password) // Password should be hashed
			} else {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	hashedPassword, _ := service.HashPassword("password123")
	validUser := &domain.User{
		Id:        uuid.New(),
		Name:      "John Doe",
		Email:     "john@example.com",
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		request       service.LoginRequest
		mockUserSetup func(*mockUserRepository)
		mockJWTSetup  func(*mockJWTProvider)
		shouldSucceed bool
		expectedError string
	}{
		{
			name: "successful login",
			request: service.LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockUserSetup: func(repo *mockUserRepository) {
				repo.On("GetByEmail", mock.Anything, "john@example.com").Return(validUser, nil)
				repo.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Run(func(args mock.Arguments) {
					token := args.Get(1).(*domain.RefreshToken)
					token.Id = uuid.New()
				})
			},
			mockJWTSetup: func(jwt *mockJWTProvider) {
				jwt.On("Generate", validUser.Id.String(), mock.AnythingOfType("time.Time"), mock.AnythingOfType("map[string]string")).Return("jwt-token", nil)
			},
			shouldSucceed: true,
		},
		{
			name: "user not found",
			request: service.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockUserSetup: func(repo *mockUserRepository) {
				repo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, domain.NotFoundError("user not found"))
			},
			mockJWTSetup:  func(jwt *mockJWTProvider) {},
			shouldSucceed: false,
			expectedError: "invalid credentials",
		},
		{
			name: "invalid password",
			request: service.LoginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			mockUserSetup: func(repo *mockUserRepository) {
				repo.On("GetByEmail", mock.Anything, "john@example.com").Return(validUser, nil)
			},
			mockJWTSetup:  func(jwt *mockJWTProvider) {},
			shouldSucceed: false,
			expectedError: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			mockJWT := &mockJWTProvider{}
			tt.mockUserSetup(mockRepo)
			tt.mockJWTSetup(mockJWT)

			service := service.NewUserService(mockJWT, mockRepo)
			ctx := context.Background()

			result, err := service.Login(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
				assert.Equal(t, validUser, result.User)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedError != "" {
					var domainErr domain.DomainError
					if errors.As(err, &domainErr) {
						assert.Contains(t, domainErr.Error(), tt.expectedError)
					}
				}
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestUserService_RefreshToken(t *testing.T) {
	validUser := &domain.User{
		Id:        uuid.New(),
		Name:      "John Doe",
		Email:     "john@example.com",
		Password:  "hashedpass",
		CreatedAt: time.Now(),
	}

	validRefreshToken := &domain.RefreshToken{
		Id:        uuid.New(),
		UserId:    validUser.Id,
		Token:     "valid-refresh-token",
		Active:    true,
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		request       service.RefreshTokenRequest
		mockUserSetup func(*mockUserRepository)
		mockJWTSetup  func(*mockJWTProvider)
		shouldSucceed bool
		expectedError string
	}{
		{
			name: "successful token refresh",
			request: service.RefreshTokenRequest{
				Token: "valid-refresh-token",
			},
			mockUserSetup: func(repo *mockUserRepository) {
				repo.On("GetRefreshToken", mock.Anything, "valid-refresh-token").Return(validRefreshToken, nil)
				repo.On("GetById", mock.Anything, validUser.Id).Return(validUser, nil)
				repo.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Run(func(args mock.Arguments) {
					token := args.Get(1).(*domain.RefreshToken)
					token.Id = uuid.New()
				})
				repo.On("UpdateRefreshTokenActive", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)
			},
			mockJWTSetup: func(jwt *mockJWTProvider) {
				jwt.On("Generate", validUser.Id.String(), mock.AnythingOfType("time.Time"), mock.AnythingOfType("map[string]string")).Return("new-jwt-token", nil)
			},
			shouldSucceed: true,
		},
		{
			name: "invalid refresh token",
			request: service.RefreshTokenRequest{
				Token: "invalid-token",
			},
			mockUserSetup: func(repo *mockUserRepository) {
				repo.On("GetRefreshToken", mock.Anything, "invalid-token").Return(nil, domain.NotFoundError("refresh token not found"))
			},
			mockJWTSetup:  func(jwt *mockJWTProvider) {},
			shouldSucceed: false,
		},
		{
			name: "inactive refresh token",
			request: service.RefreshTokenRequest{
				Token: "inactive-token",
			},
			mockUserSetup: func(repo *mockUserRepository) {
				inactiveToken := &domain.RefreshToken{
					Id:        uuid.New(),
					UserId:    validUser.Id,
					Token:     "inactive-token",
					Active:    false,
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
				}
				repo.On("GetRefreshToken", mock.Anything, "inactive-token").Return(inactiveToken, nil)
			},
			mockJWTSetup:  func(jwt *mockJWTProvider) {},
			shouldSucceed: false,
			expectedError: "invalid refresh token",
		},
		{
			name: "expired refresh token",
			request: service.RefreshTokenRequest{
				Token: "expired-token",
			},
			mockUserSetup: func(repo *mockUserRepository) {
				expiredToken := &domain.RefreshToken{
					Id:        uuid.New(),
					UserId:    validUser.Id,
					Token:     "expired-token",
					Active:    true,
					ExpiresAt: time.Now().Add(-time.Hour), // Expired
					CreatedAt: time.Now(),
				}
				repo.On("GetRefreshToken", mock.Anything, "expired-token").Return(expiredToken, nil)
			},
			mockJWTSetup:  func(jwt *mockJWTProvider) {},
			shouldSucceed: false,
			expectedError: "invalid refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			mockJWT := &mockJWTProvider{}
			tt.mockUserSetup(mockRepo)
			tt.mockJWTSetup(mockJWT)

			service := service.NewUserService(mockJWT, mockRepo)
			ctx := context.Background()

			result, err := service.RefreshToken(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
				assert.Equal(t, validUser, result.User)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedError != "" {
					var domainErr domain.DomainError
					if errors.As(err, &domainErr) {
						assert.Contains(t, domainErr.Error(), tt.expectedError)
					}
				}
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		shouldErr bool
	}{
		{"valid password", "password123", false},
		{"empty password", "", false},
		{"long password", "this-is-a-very-long-password-that-should-still-work", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := service.HashPassword(tt.password)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash)
			}
		})
	}
}

func TestCompareHash(t *testing.T) {
	password := "testpassword"
	hash, _ := service.HashPassword(password)

	tests := []struct {
		name      string
		password  string
		hash      string
		expected  bool
		shouldErr bool
	}{
		{"correct password", password, hash, true, false},
		{"incorrect password", "wrongpassword", hash, false, true},
		{"empty password", "", hash, false, true},
		{"invalid hash", password, "invalid-hash", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CompareHash(tt.password, tt.hash)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		shouldErr bool
	}{
		{"valid length", 32, false},
		{"zero length", 0, false},
		{"large length", 128, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateRefreshToken(tt.length)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				if tt.length > 0 {
					assert.NotEmpty(t, token)
				}
			}
		})
	}

	// Test uniqueness
	t.Run("generates unique tokens", func(t *testing.T) {
		token1, _ := service.GenerateRefreshToken(32)
		token2, _ := service.GenerateRefreshToken(32)
		assert.NotEqual(t, token1, token2)
	})
}
