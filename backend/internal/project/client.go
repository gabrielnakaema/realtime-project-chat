package project

import (
	"context"

	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client projectv1.ProjectServiceClient
}

func NewClient(client projectv1.ProjectServiceClient) *Client {
	return &Client{client: client}
}

// CheckAccess returns nil when the user may access the project, and a gRPC
// status error otherwise. It fails closed if the server returns an empty
// response.
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
