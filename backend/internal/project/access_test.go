package project

import (
	"context"
	"errors"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rpcClientStub struct {
	response *projectv1.CheckAccessResponse
	err      error
}

func (s rpcClientStub) CheckAccess(context.Context, *projectv1.CheckAccessRequest, ...grpc.CallOption) (*projectv1.CheckAccessResponse, error) {
	return s.response, s.err
}

type projectServiceStub struct {
	err error
}

func (s projectServiceStub) GetById(context.Context, uuid.UUID, uuid.UUID) (*domain.Project, error) {
	return &domain.Project{}, s.err
}

func TestServerAllowsAccessibleProject(t *testing.T) {
	server := NewServer(projectServiceStub{})
	response, err := server.CheckAccess(context.Background(), &projectv1.CheckAccessRequest{
		UserId: uuid.NewString(), ProjectId: uuid.NewString(),
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
			server := NewServer(projectServiceStub{err: test.err})
			_, err := server.CheckAccess(context.Background(), &projectv1.CheckAccessRequest{
				UserId: uuid.NewString(), ProjectId: uuid.NewString(),
			})
			require.Equal(t, test.code, status.Code(err))
		})
	}
}

func TestServerRejectsMalformedRequests(t *testing.T) {
	server := NewServer(projectServiceStub{})
	requests := []*projectv1.CheckAccessRequest{
		nil,
		{},
		{UserId: uuid.NewString(), ProjectId: "bad"},
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
