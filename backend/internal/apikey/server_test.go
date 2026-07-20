package apikey

import (
	"context"
	"testing"

	apikeyv1 "github.com/gabrielnakaema/project-chat/internal/apikey/v1"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authenticationServiceStub struct {
	result *AuthenticateMCPAPIKeyResult
	err    error
}

func (s authenticationServiceStub) Authenticate(context.Context, string) (*AuthenticateMCPAPIKeyResult, error) {
	return s.result, s.err
}

func TestServerReturnsSafeAuthenticationSummary(t *testing.T) {
	key := &domain.MCPAPIKey{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		SecretHash: "must-not-cross-the-process-boundary",
		Scopes:     []domain.MCPAPIScope{domain.MCPAPIScopeProjectsRead},
	}
	server := NewServer(authenticationServiceStub{result: &AuthenticateMCPAPIKeyResult{Key: key}})

	response, err := server.Authenticate(context.Background(), &apikeyv1.AuthenticateRequest{BearerSecret: "mcp_prefix_secret"})
	require.NoError(t, err)
	require.Equal(t, key.ID.String(), response.GetKeyId())
	require.Equal(t, key.UserID.String(), response.GetUserId())
	require.Equal(t, []string{string(domain.MCPAPIScopeProjectsRead)}, response.GetScopes())
}

func TestServerRejectsMissingSecretAndMapsUnauthorized(t *testing.T) {
	server := NewServer(authenticationServiceStub{})
	_, err := server.Authenticate(context.Background(), &apikeyv1.AuthenticateRequest{})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	server = NewServer(authenticationServiceStub{err: apperr.UnauthorizedError("invalid api key")})
	_, err = server.Authenticate(context.Background(), &apikeyv1.AuthenticateRequest{BearerSecret: "invalid"})
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}
