package project

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type projectReadService interface {
	GetById(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Project, error)
	ListByUserId(ctx context.Context, request ListProjectsByUserIdRequest) ([]domain.Project, error)
}

func (s *Server) GetProject(ctx context.Context, request *projectv1.GetProjectRequest) (*projectv1.GetProjectResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userID, err := requiredUUID(request.GetUserId(), "user_id")
	if err != nil {
		return nil, err
	}
	projectID, err := requiredUUID(request.GetProjectId(), "project_id")
	if err != nil {
		return nil, err
	}

	value, err := s.projectService.GetById(ctx, projectID, userID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encode project")
	}

	return &projectv1.GetProjectResponse{ProjectJson: payload}, nil
}

func (s *Server) ListProjects(ctx context.Context, request *projectv1.ListProjectsRequest) (*projectv1.ListProjectsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userID, err := requiredUUID(request.GetUserId(), "user_id")
	if err != nil {
		return nil, err
	}

	values, err := s.projectService.ListByUserId(ctx, ListProjectsByUserIdRequest{
		UserId:             userID,
		MemberRole:         domain.ProjectMemberRole(request.GetMemberRole()),
		ShouldFilterByRole: request.GetFilterByRole(),
		SearchQuery:        request.GetSearchQuery(),
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encode projects")
	}

	return &projectv1.ListProjectsResponse{ProjectsJson: payload}, nil
}

func requiredUUID(value string, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, status.Errorf(codes.InvalidArgument, "valid %s is required", field)
	}
	return id, nil
}

type Server struct {
	projectv1.UnimplementedProjectServiceServer
	projectService projectReadService
}

func NewServer(projectService projectReadService) *Server {
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
	var domainError apperr.DomainError
	if !errors.As(err, &domainError) {
		return status.Error(codes.Internal, "project access check failed")
	}

	switch domainError.Code {
	case apperr.NotFoundErrorCode:
		return status.Error(codes.NotFound, "project not found")
	case apperr.ForbiddenErrorCode, apperr.UnauthorizedErrorCode:
		return status.Error(codes.PermissionDenied, "project access forbidden")
	case apperr.ValidationFailedErrorCode, apperr.BusinessValidationErrorCode:
		return status.Error(codes.InvalidArgument, "project access check request is invalid")
	default:
		return status.Error(codes.Internal, "project access check failed")
	}
}
