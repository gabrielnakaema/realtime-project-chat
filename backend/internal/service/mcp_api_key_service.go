package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
)

const mcpAPIKeyLastUsedWriteInterval = 5 * time.Minute

type mcpAPIKeyRepository interface {
	Create(ctx context.Context, key *domain.MCPAPIKey) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.MCPAPIKey, error)
	GetByIDForUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.MCPAPIKey, error)
	GetByPrefix(ctx context.Context, prefix string) (*domain.MCPAPIKey, error)
	Revoke(ctx context.Context, id uuid.UUID, userID uuid.UUID, revokedAt time.Time) error
	TouchLastUsedAt(ctx context.Context, id uuid.UUID, now time.Time, minInterval time.Duration) error
}

type MCPAPIKeyService struct {
	repository mcpAPIKeyRepository
}

func NewMCPAPIKeyService(repository mcpAPIKeyRepository) *MCPAPIKeyService {
	return &MCPAPIKeyService{repository: repository}
}

type CreateMCPAPIKeyRequest struct {
	UserID uuid.UUID
	Name   string
	Scopes []domain.MCPAPIScope
}

type CreateMCPAPIKeyResult struct {
	Key       *domain.MCPAPIKey `json:"key"`
	RawSecret string            `json:"raw_secret"`
}

func (s *MCPAPIKeyService) ListAvailableScopes() []domain.MCPAPIScopeDefinition {
	scopes := make([]domain.MCPAPIScopeDefinition, len(domain.MCPAPIScopeDefinitions))
	copy(scopes, domain.MCPAPIScopeDefinitions)

	return scopes
}

func (s *MCPAPIKeyService) Create(ctx context.Context, request CreateMCPAPIKeyRequest) (*CreateMCPAPIKeyResult, error) {
	if request.UserID == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	name := strings.TrimSpace(request.Name)
	if name == "" {
		return nil, domain.BusinessValidationError("name is required")
	}

	scopes, err := normalizeMCPScopes(request.Scopes)
	if err != nil {
		return nil, err
	}

	prefix, err := generateHexSecret(4)
	if err != nil {
		return nil, domain.ServerError("failed to generate mcp api key", err)
	}

	rawToken, err := generateOpaqueSecret(32)
	if err != nil {
		return nil, domain.ServerError("failed to generate mcp api key", err)
	}

	rawSecret := "mcp_" + prefix + "_" + rawToken
	now := time.Now()

	key := &domain.MCPAPIKey{
		UserID:     request.UserID,
		Name:       name,
		KeyPrefix:  prefix,
		SecretHash: hashMCPSecret(rawToken),
		Scopes:     scopes,
		CreatedAt:  now,
	}

	if err := s.repository.Create(ctx, key); err != nil {
		return nil, domain.ServerError("failed to create mcp api key", err)
	}

	return &CreateMCPAPIKeyResult{
		Key:       key,
		RawSecret: rawSecret,
	}, nil
}

func (s *MCPAPIKeyService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.MCPAPIKey, error) {
	if userID == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	keys, err := s.repository.ListByUserID(ctx, userID)
	if err != nil {
		return nil, domain.ServerError("failed to list mcp api keys", err)
	}

	return keys, nil
}

func (s *MCPAPIKeyService) Revoke(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return domain.UnauthorizedError("unauthorized")
	}

	key, err := s.repository.GetByIDForUser(ctx, id, userID)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			return domainErr
		}
		return domain.ServerError("failed to get mcp api key", err)
	}

	if key.IsRevoked() {
		return domain.BusinessValidationError("mcp api key is already revoked")
	}

	if err := s.repository.Revoke(ctx, id, userID, time.Now()); err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			return domainErr
		}
		return domain.ServerError("failed to revoke mcp api key", err)
	}

	return nil
}

type AuthenticateMCPAPIKeyResult struct {
	Key *domain.MCPAPIKey
}

func (s *MCPAPIKeyService) Authenticate(ctx context.Context, bearerSecret string) (*AuthenticateMCPAPIKeyResult, error) {
	prefix, token, err := parseMCPSecret(bearerSecret)
	if err != nil {
		return nil, domain.UnauthorizedError("invalid api key")
	}

	key, err := s.repository.GetByPrefix(ctx, prefix)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) && domainErr.Code == domain.NotFoundErrorCode {
			return nil, domain.UnauthorizedError("invalid api key")
		}
		if errors.As(err, &domainErr) {
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to load mcp api key", err)
	}

	expected := []byte(key.SecretHash)
	actual := []byte(hashMCPSecret(token))
	if subtle.ConstantTimeCompare(expected, actual) != 1 {
		return nil, domain.UnauthorizedError("invalid api key")
	}

	if key.IsRevoked() {
		return nil, domain.UnauthorizedError("api key revoked")
	}

	if err := s.repository.TouchLastUsedAt(ctx, key.ID, time.Now(), mcpAPIKeyLastUsedWriteInterval); err != nil {
		return nil, domain.ServerError("failed to update mcp api key usage", err)
	}

	return &AuthenticateMCPAPIKeyResult{Key: key}, nil
}

func normalizeMCPScopes(scopes []domain.MCPAPIScope) ([]domain.MCPAPIScope, error) {
	if len(scopes) == 0 {
		return nil, domain.BusinessValidationError("at least one scope is required")
	}

	key := domain.MCPAPIKey{}
	unique := map[domain.MCPAPIScope]struct{}{}
	normalized := make([]domain.MCPAPIScope, 0, len(scopes))
	for _, scope := range scopes {
		scope = domain.MCPAPIScope(strings.TrimSpace(string(scope)))
		if !key.IsAllowedScope(scope) {
			return nil, domain.BusinessValidationError("invalid mcp api key scope")
		}
		if _, ok := unique[scope]; ok {
			continue
		}
		unique[scope] = struct{}{}
		normalized = append(normalized, scope)
	}

	return normalized, nil
}

func parseMCPSecret(raw string) (prefix string, token string, err error) {
	parts := strings.SplitN(strings.TrimSpace(raw), "_", 3)
	if len(parts) != 3 || parts[0] != "mcp" || parts[1] == "" || parts[2] == "" {
		return "", "", errors.New("invalid mcp api key")
	}

	return parts[1], parts[2], nil
}

func hashMCPSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func generateOpaqueSecret(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func generateHexSecret(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
