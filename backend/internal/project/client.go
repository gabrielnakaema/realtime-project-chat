package project

import (
	"context"
	"encoding/json"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListRequest struct {
	UserID             uuid.UUID
	MemberRole         domain.ProjectMemberRole
	ShouldFilterByRole bool
	SearchQuery        string
}

type Client struct {
	client projectv1.ProjectServiceClient
}

func NewClient(client projectv1.ProjectServiceClient) *Client {
	return &Client{client: client}
}

func (c *Client) CheckAccess(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) error {
	response, err := c.client.CheckAccess(ctx, &projectv1.CheckAccessRequest{
		UserId:    userID.String(),
		ProjectId: projectID.String(),
	})
	if err != nil {
		return err
	}
	if response == nil {
		return status.Error(codes.Internal, "empty project access check response")
	}

	return nil
}

func (c *Client) Get(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (*domain.Project, error) {
	response, err := c.client.GetProject(ctx, &projectv1.GetProjectRequest{
		UserId:    userID.String(),
		ProjectId: projectID.String(),
	})
	if err != nil {
		return nil, projectStatusToDomainError(err)
	}
	if response == nil {
		return nil, apperr.ServerError("project service returned an empty response", nil)
	}

	var value domain.Project
	if err := json.Unmarshal(response.GetProjectJson(), &value); err != nil {
		return nil, apperr.ServerError("failed to decode project", err)
	}
	return &value, nil
}

func (c *Client) List(ctx context.Context, request ListRequest) ([]domain.Project, error) {
	response, err := c.client.ListProjects(ctx, &projectv1.ListProjectsRequest{
		UserId:       request.UserID.String(),
		MemberRole:   string(request.MemberRole),
		FilterByRole: request.ShouldFilterByRole,
		SearchQuery:  request.SearchQuery,
	})
	if err != nil {
		return nil, projectStatusToDomainError(err)
	}
	if response == nil {
		return nil, apperr.ServerError("project service returned an empty response", nil)
	}

	var values []domain.Project
	if err := json.Unmarshal(response.GetProjectsJson(), &values); err != nil {
		return nil, apperr.ServerError("failed to decode projects", err)
	}
	return values, nil
}

func projectStatusToDomainError(err error) error {
	value, ok := status.FromError(err)
	if !ok {
		return apperr.ServerError("project service call failed", err)
	}

	switch value.Code() {
	case codes.NotFound:
		return apperr.NotFoundError(value.Message())
	case codes.Unauthenticated:
		return apperr.UnauthorizedError(value.Message())
	case codes.PermissionDenied:
		return apperr.ForbiddenError(value.Message())
	case codes.InvalidArgument:
		return apperr.BusinessValidationError(value.Message())
	default:
		return apperr.ServerError(value.Message(), err)
	}
}
