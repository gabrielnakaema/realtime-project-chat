package ws

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/api"
	"github.com/gabrielnakaema/project-chat/internal/chat"
	chatv1 "github.com/gabrielnakaema/project-chat/internal/chat/v1"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/project"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type accessCheckerStub struct {
	check func(context.Context, uuid.UUID, uuid.UUID) error
}

func (s accessCheckerStub) CheckAccess(ctx context.Context, userID, resourceID uuid.UUID) error {
	if s.check == nil {
		return nil
	}
	return s.check(ctx, userID, resourceID)
}

// failIfCalled fails the test if the checker is ever invoked.
func failIfCalled(t *testing.T) accessCheckerStub {
	return accessCheckerStub{check: func(context.Context, uuid.UUID, uuid.UUID) error {
		t.Helper()
		t.Fatal("access checker unexpectedly called")
		return nil
	}}
}

type testPublisher struct {
	published chan events.Topic
}

func (p testPublisher) Publish(_ context.Context, topic events.Topic, _ events.Payload) error {
	if p.published != nil {
		p.published <- topic
	}
	return nil
}

func newRoomTestServer(t *testing.T, chatAuthorizer, projectAuthorizer accessChecker, timeout time.Duration, publisher publisher) (*Server, uuid.UUID) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	server := NewServer(ctx, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), chatAuthorizer, projectAuthorizer, timeout, publisher)
	userID := uuid.New()
	server.users[userID] = &WsUser{id: userID, rooms: map[uuid.UUID]bool{}, writer: make(chan any, 4)}
	return server, userID
}

func TestConnectUserToRoomAuthorizesBeforeMutationAndPublication(t *testing.T) {
	type contextKey string
	const requestKey contextKey = "request"
	authorized := false
	published := make(chan events.Topic, 1)
	chatAuthorizer := accessCheckerStub{check: func(ctx context.Context, _ uuid.UUID, _ uuid.UUID) error {
		require.Equal(t, "connection-context", ctx.Value(requestKey))
		authorized = true
		return nil
	}}
	server, userID := newRoomTestServer(t, chatAuthorizer, failIfCalled(t), time.Second, testPublisher{published: published})
	roomID := uuid.New()

	ctx := context.WithValue(context.Background(), requestKey, "connection-context")
	require.NoError(t, server.connectUserToRoom(ctx, userID, roomID, WsRoomTypeChat))
	require.True(t, authorized)
	require.True(t, server.users[userID].rooms[roomID])
	require.True(t, server.rooms[roomID].users[userID])
	require.Equal(t, events.ChatMemberViewed, <-published)
}

func TestUserRoomAuthorizationRemainsLocal(t *testing.T) {
	server, userID := newRoomTestServer(t, failIfCalled(t), failIfCalled(t), time.Second, testPublisher{})

	err := server.connectUserToRoom(context.Background(), userID, uuid.New(), WsRoomTypeUser)
	require.Error(t, err)
	require.Empty(t, server.rooms)
}

func TestConnectUserToRoomFailsClosedForAuthorizationErrors(t *testing.T) {
	for _, code := range []codes.Code{codes.PermissionDenied, codes.NotFound, codes.Unavailable, codes.InvalidArgument, codes.Internal} {
		t.Run(code.String(), func(t *testing.T) {
			projectAuthorizer := accessCheckerStub{check: func(context.Context, uuid.UUID, uuid.UUID) error {
				return status.Error(code, "denied")
			}}
			server, userID := newRoomTestServer(t, failIfCalled(t), projectAuthorizer, time.Second, testPublisher{})
			roomID := uuid.New()

			err := server.connectUserToRoom(context.Background(), userID, roomID, WsRoomTypeProject)
			require.Equal(t, code, status.Code(err))
			require.NotContains(t, server.users[userID].rooms, roomID)
			require.NotContains(t, server.rooms, roomID)
		})
	}
}

func TestDeniedConnectionDoesNotEmitUserConnected(t *testing.T) {
	chatAuthorizer := accessCheckerStub{check: func(context.Context, uuid.UUID, uuid.UUID) error {
		return status.Error(codes.Unavailable, "authorization unavailable")
	}}
	server, userID := newRoomTestServer(t, chatAuthorizer, failIfCalled(t), time.Second, testPublisher{})
	roomID := uuid.New()
	writer := server.users[userID].writer

	server.handleConnectUserToRoom(context.Background(), userID, WebsocketMessage{
		Type: WebsocketMessageTypeConnectUserToRoom,
		Data: ConnectUserToRoomData{RoomId: roomID, Type: WsRoomTypeChat},
	})

	require.NotContains(t, server.rooms, roomID)
	select {
	case message := <-writer:
		t.Fatalf("unexpected successful connection message: %#v", message)
	default:
	}
}

func TestConnectUserToRoomAppliesAuthorizationDeadline(t *testing.T) {
	chatAuthorizer := accessCheckerStub{check: func(ctx context.Context, _ uuid.UUID, _ uuid.UUID) error {
		<-ctx.Done()
		return status.Error(codes.DeadlineExceeded, ctx.Err().Error())
	}}
	server, userID := newRoomTestServer(t, chatAuthorizer, failIfCalled(t), 10*time.Millisecond, testPublisher{})
	roomID := uuid.New()

	err := server.connectUserToRoom(context.Background(), userID, roomID, WsRoomTypeChat)
	require.Equal(t, codes.DeadlineExceeded, status.Code(err))
	require.NotContains(t, server.rooms, roomID)
}

type splitChatService struct {
	allowed bool
}

func (s splitChatService) GetById(context.Context, uuid.UUID, uuid.UUID) (*domain.Chat, error) {
	if !s.allowed {
		return nil, domain.ForbiddenError("forbidden")
	}
	return &domain.Chat{}, nil
}

type splitProjectService struct {
	allowed bool
}

func (s splitProjectService) GetById(context.Context, uuid.UUID, uuid.UUID) (*domain.Project, error) {
	if !s.allowed {
		return nil, domain.ForbiddenError("forbidden")
	}
	return &domain.Project{}, nil
}

// TestSplitServiceRoomAuthorization exercises the full gRPC path: the WS server
// routes each room type to its domain client, which calls the matching domain
// server registered on the core service.
func TestSplitServiceRoomAuthorization(t *testing.T) {
	for _, test := range []struct {
		name           string
		roomType       WsRoomType
		chatAllowed    bool
		projectAllowed bool
		wantCode       codes.Code
	}{
		{name: "allowed chat", roomType: WsRoomTypeChat, chatAllowed: true, wantCode: codes.OK},
		{name: "denied chat", roomType: WsRoomTypeChat, wantCode: codes.PermissionDenied},
		{name: "allowed project", roomType: WsRoomTypeProject, projectAllowed: true, wantCode: codes.OK},
		{name: "denied project", roomType: WsRoomTypeProject, wantCode: codes.PermissionDenied},
	} {
		t.Run(test.name, func(t *testing.T) {
			listener := bufconn.Listen(1024 * 1024)
			grpcServer := api.NewInternalGRPCServer(
				chat.NewServer(splitChatService{allowed: test.chatAllowed}),
				project.NewServer(splitProjectService{allowed: test.projectAllowed}),
			)
			go func() { _ = grpcServer.Serve(listener) }()
			t.Cleanup(grpcServer.Stop)

			connection, err := grpc.NewClient("passthrough:///core", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return listener.Dial()
			}))
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, connection.Close()) })

			chatAuthorizer := chat.NewClient(chatv1.NewChatServiceClient(connection))
			projectAuthorizer := project.NewClient(projectv1.NewProjectServiceClient(connection))
			server, userID := newRoomTestServer(t, chatAuthorizer, projectAuthorizer, time.Second, testPublisher{})
			err = server.connectUserToRoom(context.Background(), userID, uuid.New(), test.roomType)
			require.Equal(t, test.wantCode, status.Code(err))
			require.Equal(t, test.wantCode == codes.OK, len(server.rooms) == 1)
		})
	}
}
