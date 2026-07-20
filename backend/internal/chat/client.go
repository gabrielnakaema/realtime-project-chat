package chat

import (
	"context"

	chatv1 "github.com/gabrielnakaema/project-chat/internal/chat/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client chatv1.ChatServiceClient
}

func NewClient(client chatv1.ChatServiceClient) *Client {
	return &Client{client: client}
}

// CheckAccess returns nil when the user may access the chat, and a gRPC status
// error otherwise. It fails closed if the server returns an empty response.
func (c *Client) CheckAccess(ctx context.Context, userID uuid.UUID, chatID uuid.UUID) error {
	response, err := c.client.CheckAccess(ctx, &chatv1.CheckAccessRequest{
		UserId: userID.String(),
		ChatId: chatID.String(),
	})
	if err != nil {
		return err
	}
	if response == nil {
		return status.Error(codes.Internal, "empty chat access check response")
	}

	return nil
}
