package apikey

import (
	"context"

	apikeyv1 "github.com/gabrielnakaema/project-chat/internal/apikey/v1"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthenticatedKey struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Scopes []domain.MCPAPIScope
}

type Client struct {
	client apikeyv1.APIKeyServiceClient
}

func NewClient(client apikeyv1.APIKeyServiceClient) *Client {
	return &Client{client: client}
}

func (c *Client) Authenticate(ctx context.Context, bearerSecret string) (*AuthenticatedKey, error) {
	response, err := c.client.Authenticate(ctx, &apikeyv1.AuthenticateRequest{BearerSecret: bearerSecret})
	if err != nil {
		return nil, statusToDomainError(err)
	}
	if response == nil {
		return nil, apperr.ServerError("api key authentication returned an empty response", nil)
	}

	keyID, err := uuid.Parse(response.GetKeyId())
	if err != nil || keyID == uuid.Nil {
		return nil, apperr.ServerError("api key authentication returned an invalid key id", err)
	}
	userID, err := uuid.Parse(response.GetUserId())
	if err != nil || userID == uuid.Nil {
		return nil, apperr.ServerError("api key authentication returned an invalid user id", err)
	}

	scopes := make([]domain.MCPAPIScope, 0, len(response.GetScopes()))
	for _, scope := range response.GetScopes() {
		scopes = append(scopes, domain.MCPAPIScope(scope))
	}

	return &AuthenticatedKey{ID: keyID, UserID: userID, Scopes: scopes}, nil
}

func statusToDomainError(err error) error {
	statusError, ok := status.FromError(err)
	if !ok {
		return apperr.ServerError("api key service call failed", err)
	}

	switch statusError.Code() {
	case codes.Unauthenticated:
		return apperr.UnauthorizedError(statusError.Message())
	case codes.InvalidArgument:
		return apperr.BusinessValidationError(statusError.Message())
	default:
		return apperr.ServerError(statusError.Message(), err)
	}
}
