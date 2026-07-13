package project

import (
	"context"
	"errors"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type projectService interface {
	GetById(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Project, error)
}

type Server struct {
	projectv1.UnimplementedProjectServiceServer
	projectService projectService
}

func NewServer(projectService projectService) *Server {
	return &Server{projectService: projectService}
}

func (s *Server) CheckAccess(ctx context.Context, request *projectv1.CheckAccessRequest) (*projectv1.CheckAccessResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userID, err := uuid.Parse(request.GetUserId())
	if err != nil || userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}
	projectID, err := uuid.Parse(request.GetProjectId())
	if err != nil || projectID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "valid project_id is required")
	}

	if _, err := s.projectService.GetById(ctx, projectID, userID); err != nil {
		return nil, mapDomainError(err)
	}

	return &projectv1.CheckAccessResponse{}, nil
}

func mapDomainError(err error) error {
	var domainError domain.DomainError
	if !errors.As(err, &domainError) {
		return status.Error(codes.Internal, "project access check failed")
	}

	switch domainError.Code {
	case domain.NotFoundErrorCode:
		return status.Error(codes.NotFound, "project not found")
	case domain.ForbiddenErrorCode, domain.UnauthorizedErrorCode:
		return status.Error(codes.PermissionDenied, "project access forbidden")
	case domain.ValidationFailedErrorCode, domain.BusinessValidationErrorCode:
		return status.Error(codes.InvalidArgument, "project access check request is invalid")
	default:
		return status.Error(codes.Internal, "project access check failed")
	}
}
