package apikey

import (
	"context"
	"errors"
	"strings"

	apikeyv1 "github.com/gabrielnakaema/project-chat/internal/apikey/v1"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authenticationService interface {
	Authenticate(context.Context, string) (*AuthenticateMCPAPIKeyResult, error)
}

type Server struct {
	apikeyv1.UnimplementedAPIKeyServiceServer
	service authenticationService
}

func NewServer(service authenticationService) *Server {
	return &Server{service: service}
}

func (s *Server) Authenticate(ctx context.Context, request *apikeyv1.AuthenticateRequest) (*apikeyv1.AuthenticateResponse, error) {
	if request == nil || strings.TrimSpace(request.GetBearerSecret()) == "" {
		return nil, status.Error(codes.InvalidArgument, "bearer_secret is required")
	}

	result, err := s.service.Authenticate(ctx, request.GetBearerSecret())
	if err != nil {
		return nil, mapDomainError(err)
	}
	if result == nil || result.Key == nil {
		return nil, status.Error(codes.Internal, "api key authentication returned an empty result")
	}

	scopes := make([]string, 0, len(result.Key.Scopes))
	for _, scope := range result.Key.Scopes {
		scopes = append(scopes, string(scope))
	}

	return &apikeyv1.AuthenticateResponse{
		KeyId:  result.Key.ID.String(),
		UserId: result.Key.UserID.String(),
		Scopes: scopes,
	}, nil
}

func mapDomainError(err error) error {
	var domainError apperr.DomainError
	if !errors.As(err, &domainError) {
		return status.Error(codes.Internal, "api key authentication failed")
	}

	switch domainError.Code {
	case apperr.UnauthorizedErrorCode:
		return status.Error(codes.Unauthenticated, domainError.Message)
	case apperr.ValidationFailedErrorCode, apperr.BusinessValidationErrorCode:
		return status.Error(codes.InvalidArgument, domainError.Message)
	default:
		return status.Error(codes.Internal, "api key authentication failed")
	}
}
