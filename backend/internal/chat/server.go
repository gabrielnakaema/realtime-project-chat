package chat

import (
	"context"
	"errors"

	chatv1 "github.com/gabrielnakaema/project-chat/internal/chat/v1"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accessService interface {
	GetById(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Chat, error)
}

type Server struct {
	chatv1.UnimplementedChatServiceServer
	chatService accessService
}

func NewServer(chatService accessService) *Server {
	return &Server{chatService: chatService}
}

func (s *Server) CheckAccess(ctx context.Context, request *chatv1.CheckAccessRequest) (*chatv1.CheckAccessResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userID, err := uuid.Parse(request.GetUserId())
	if err != nil || userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}
	chatID, err := uuid.Parse(request.GetChatId())
	if err != nil || chatID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "valid chat_id is required")
	}

	if _, err := s.chatService.GetById(ctx, chatID, userID); err != nil {
		return nil, mapDomainError(err)
	}

	return &chatv1.CheckAccessResponse{}, nil
}

func mapDomainError(err error) error {
	var domainError domain.DomainError
	if !errors.As(err, &domainError) {
		return status.Error(codes.Internal, "chat access check failed")
	}

	switch domainError.Code {
	case domain.NotFoundErrorCode:
		return status.Error(codes.NotFound, "chat not found")
	case domain.ForbiddenErrorCode, domain.UnauthorizedErrorCode:
		return status.Error(codes.PermissionDenied, "chat access forbidden")
	case domain.ValidationFailedErrorCode, domain.BusinessValidationErrorCode:
		return status.Error(codes.InvalidArgument, "chat access check request is invalid")
	default:
		return status.Error(codes.Internal, "chat access check failed")
	}
}
