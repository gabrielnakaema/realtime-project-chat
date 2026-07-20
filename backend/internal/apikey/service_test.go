package apikey_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/apikey"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMCPAPIKeyRepository struct {
	mock.Mock
}

func (m *mockMCPAPIKeyRepository) Create(ctx context.Context, key *domain.MCPAPIKey) error {
	return m.Called(ctx, key).Error(0)
}

func (m *mockMCPAPIKeyRepository) Update(ctx context.Context, key *domain.MCPAPIKey) error {
	return m.Called(ctx, key).Error(0)
}

func (m *mockMCPAPIKeyRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.MCPAPIKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MCPAPIKey), args.Error(1)
}

func (m *mockMCPAPIKeyRepository) GetByIDForUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.MCPAPIKey, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MCPAPIKey), args.Error(1)
}

func (m *mockMCPAPIKeyRepository) GetByPrefix(ctx context.Context, prefix string) (*domain.MCPAPIKey, error) {
	args := m.Called(ctx, prefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MCPAPIKey), args.Error(1)
}

func (m *mockMCPAPIKeyRepository) Revoke(ctx context.Context, id uuid.UUID, userID uuid.UUID, revokedAt time.Time) error {
	return m.Called(ctx, id, userID, revokedAt).Error(0)
}

func (m *mockMCPAPIKeyRepository) TouchLastUsedAt(ctx context.Context, id uuid.UUID, now time.Time, minInterval time.Duration) error {
	return m.Called(ctx, id, now, minInterval).Error(0)
}

func TestMCPAPIKeyService_CreateNormalizesNameAndScopes(t *testing.T) {
	repo := &mockMCPAPIKeyRepository{}
	svc := apikey.NewMCPAPIKeyService(repo)
	userID := uuid.New()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.MCPAPIKey")).
		Return(nil).
		Run(func(args mock.Arguments) {
			key := args.Get(1).(*domain.MCPAPIKey)
			key.ID = uuid.New()
			assert.Equal(t, "Claude Desktop", key.Name)
			assert.Equal(t, []domain.MCPAPIScope{
				domain.MCPAPIScopeTasksRead,
				domain.MCPAPIScopeTasksComment,
			}, key.Scopes)
		})

	result, err := svc.Create(context.Background(), apikey.CreateMCPAPIKeyRequest{
		UserID: userID,
		Name:   "  Claude Desktop  ",
		Scopes: []domain.MCPAPIScope{
			domain.MCPAPIScopeTasksRead,
			domain.MCPAPIScopeTasksComment,
			domain.MCPAPIScopeTasksRead,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Claude Desktop", result.Key.Name)
	assert.Equal(t, []domain.MCPAPIScope{
		domain.MCPAPIScopeTasksRead,
		domain.MCPAPIScopeTasksComment,
	}, result.Key.Scopes)
	assert.NotEmpty(t, result.RawSecret)
	repo.AssertExpectations(t)
}

func TestMCPAPIKeyService_Update(t *testing.T) {
	t.Run("normalizes replacement fields and returns reloaded key", func(t *testing.T) {
		repo := &mockMCPAPIKeyRepository{}
		svc := apikey.NewMCPAPIKeyService(repo)
		keyID := uuid.New()
		userID := uuid.New()

		existing := &domain.MCPAPIKey{
			ID:     keyID,
			UserID: userID,
			Name:   "Old name",
			Scopes: []domain.MCPAPIScope{domain.MCPAPIScopeTasksRead},
		}
		updated := &domain.MCPAPIKey{
			ID:     keyID,
			UserID: userID,
			Name:   "Updated agent",
			Scopes: []domain.MCPAPIScope{
				domain.MCPAPIScopeProjectsRead,
				domain.MCPAPIScopeTasksMove,
			},
		}

		repo.On("GetByIDForUser", mock.Anything, keyID, userID).
			Return(existing, nil).
			Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.MCPAPIKey")).
			Return(nil).
			Run(func(args mock.Arguments) {
				key := args.Get(1).(*domain.MCPAPIKey)
				assert.Equal(t, "Updated agent", key.Name)
				assert.Equal(t, []domain.MCPAPIScope{
					domain.MCPAPIScopeProjectsRead,
					domain.MCPAPIScopeTasksMove,
				}, key.Scopes)
			}).
			Once()
		repo.On("GetByIDForUser", mock.Anything, keyID, userID).
			Return(updated, nil).
			Once()

		result, err := svc.Update(context.Background(), apikey.UpdateMCPAPIKeyRequest{
			ID:     keyID,
			UserID: userID,
			Name:   "  Updated agent  ",
			Scopes: []domain.MCPAPIScope{
				domain.MCPAPIScopeProjectsRead,
				domain.MCPAPIScopeTasksMove,
				domain.MCPAPIScopeTasksMove,
			},
		})

		require.NoError(t, err)
		assert.Equal(t, updated, result)
		repo.AssertExpectations(t)
	})

	t.Run("rejects revoked keys", func(t *testing.T) {
		repo := &mockMCPAPIKeyRepository{}
		svc := apikey.NewMCPAPIKeyService(repo)
		keyID := uuid.New()
		userID := uuid.New()
		revokedAt := time.Now()

		repo.On("GetByIDForUser", mock.Anything, keyID, userID).Return(&domain.MCPAPIKey{
			ID:        keyID,
			UserID:    userID,
			RevokedAt: &revokedAt,
		}, nil)

		result, err := svc.Update(context.Background(), apikey.UpdateMCPAPIKeyRequest{
			ID:     keyID,
			UserID: userID,
			Name:   "Updated agent",
			Scopes: []domain.MCPAPIScope{domain.MCPAPIScopeTasksRead},
		})

		require.Error(t, err)
		assert.Nil(t, result)
		var domainErr apperr.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, apperr.BusinessValidationErrorCode, domainErr.Code)
		assert.Equal(t, "mcp api key is already revoked", domainErr.Message)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("surfaces repository lookup failures", func(t *testing.T) {
		repo := &mockMCPAPIKeyRepository{}
		svc := apikey.NewMCPAPIKeyService(repo)
		keyID := uuid.New()
		userID := uuid.New()

		repo.On("GetByIDForUser", mock.Anything, keyID, userID).Return(nil, errors.New("db unavailable"))

		result, err := svc.Update(context.Background(), apikey.UpdateMCPAPIKeyRequest{
			ID:     keyID,
			UserID: userID,
			Name:   "Updated agent",
			Scopes: []domain.MCPAPIScope{domain.MCPAPIScopeTasksRead},
		})

		require.Error(t, err)
		assert.Nil(t, result)
		var domainErr apperr.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, apperr.ServerErrorCode, domainErr.Code)
		assert.Equal(t, "failed to get mcp api key", domainErr.Message)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})
}
