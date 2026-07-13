package chat

import (
	"context"
	"errors"
	"testing"

	chatv1 "github.com/gabrielnakaema/project-chat/internal/chat/v1"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rpcClientStub struct {
	response *chatv1.CheckAccessResponse
	err      error
}

func (s rpcClientStub) CheckAccess(context.Context, *chatv1.CheckAccessRequest, ...grpc.CallOption) (*chatv1.CheckAccessResponse, error) {
	return s.response, s.err
}

type chatServiceStub struct {
	err error
}

func (s chatServiceStub) GetById(context.Context, uuid.UUID, uuid.UUID) (*domain.Chat, error) {
	return &domain.Chat{}, s.err
}

func TestServerAllowsAccessibleChat(t *testing.T) {
	server := NewServer(chatServiceStub{})
	response, err := server.CheckAccess(context.Background(), &chatv1.CheckAccessRequest{
		UserId: uuid.NewString(), ChatId: uuid.NewString(),
	})
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestServerMapsDomainErrorsDeterministically(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "forbidden", err: domain.ForbiddenError("forbidden"), code: codes.PermissionDenied},
		{name: "not found", err: domain.NotFoundError("not found"), code: codes.NotFound},
		{name: "validation", err: domain.BusinessValidationError("invalid"), code: codes.InvalidArgument},
		{name: "server", err: domain.ServerError("failed", errors.New("database")), code: codes.Internal},
		{name: "unknown", err: errors.New("unknown"), code: codes.Internal},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(chatServiceStub{err: test.err})
			_, err := server.CheckAccess(context.Background(), &chatv1.CheckAccessRequest{
				UserId: uuid.NewString(), ChatId: uuid.NewString(),
			})
			require.Equal(t, test.code, status.Code(err))
		})
	}
}

func TestServerRejectsMalformedRequests(t *testing.T) {
	server := NewServer(chatServiceStub{})
	requests := []*chatv1.CheckAccessRequest{
		nil,
		{},
		{UserId: uuid.NewString(), ChatId: "bad"},
		{UserId: uuid.NewString()},
	}
	for _, request := range requests {
		_, err := server.CheckAccess(context.Background(), request)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}
}

func TestClientFailsClosedForEmptyResponse(t *testing.T) {
	client := NewClient(rpcClientStub{response: nil})
	err := client.CheckAccess(context.Background(), uuid.New(), uuid.New())
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestClientPropagatesRPCFailures(t *testing.T) {
	for _, code := range []codes.Code{codes.PermissionDenied, codes.NotFound, codes.DeadlineExceeded, codes.Unavailable, codes.InvalidArgument, codes.Internal} {
		client := NewClient(rpcClientStub{err: status.Error(code, "access check failed")})
		err := client.CheckAccess(context.Background(), uuid.New(), uuid.New())
		require.Equal(t, code, status.Code(err))
	}
}
