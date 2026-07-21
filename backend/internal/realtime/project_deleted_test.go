package realtime

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDisconnectAllFromRoomRemovesEveryMember(t *testing.T) {
	server, firstUserID := newRoomTestServer(t, failIfCalled(t), accessCheckerStub{}, time.Second, testEnqueuer{})
	secondUserID := uuid.New()
	server.users[secondUserID] = &WsUser{id: secondUserID, rooms: map[uuid.UUID]bool{}, writer: make(chan any, 4)}
	roomID := uuid.New()

	require.NoError(t, server.connectUserToRoom(context.Background(), firstUserID, roomID, WsRoomTypeProject))
	require.NoError(t, server.connectUserToRoom(context.Background(), secondUserID, roomID, WsRoomTypeProject))
	require.True(t, server.rooms[roomID].users[firstUserID])
	require.True(t, server.rooms[roomID].users[secondUserID])

	server.disconnectAllFromRoom(roomID)

	require.NotContains(t, server.rooms, roomID)
	require.NotContains(t, server.users[firstUserID].rooms, roomID)
	require.NotContains(t, server.users[secondUserID].rooms, roomID)
}

func TestDisconnectAllFromRoomIsNoOpForUnknownRoom(t *testing.T) {
	server, _ := newRoomTestServer(t, failIfCalled(t), accessCheckerStub{}, time.Second, testEnqueuer{})

	require.NotPanics(t, func() {
		server.disconnectAllFromRoom(uuid.New())
	})
}

func TestSendProjectDeletedBroadcastsAndDisconnectsRoom(t *testing.T) {
	server, userID := newRoomTestServer(t, failIfCalled(t), accessCheckerStub{}, time.Second, testEnqueuer{})
	roomID := uuid.New()

	require.NoError(t, server.connectUserToRoom(context.Background(), userID, roomID, WsRoomTypeProject))

	project := &domain.Project{Id: roomID, Name: "Doomed Project"}
	require.NoError(t, server.SendProjectDeleted(context.Background(), project))

	select {
	case message := <-server.users[userID].writer:
		wsMessage, ok := message.(WebsocketMessage)
		require.True(t, ok)
		require.Equal(t, WebsocketMessageTypeProjectDeleted, wsMessage.Type)
		require.Equal(t, roomID, wsMessage.RoomId)
		require.Equal(t, project, wsMessage.Data)
	default:
		t.Fatal("expected project_deleted message to be delivered")
	}

	require.NotContains(t, server.rooms, roomID)
	require.NotContains(t, server.users[userID].rooms, roomID)
}
