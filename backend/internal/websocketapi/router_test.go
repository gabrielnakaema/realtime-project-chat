package websocketapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gabrielnakaema/project-chat/internal/apphost"
	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/outbox"
	"github.com/gabrielnakaema/project-chat/internal/token"
	"github.com/gabrielnakaema/project-chat/internal/ws"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type accessCheckerStub struct{}

func (accessCheckerStub) CheckAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type enqueuerStub struct{}

func (enqueuerStub) Enqueue(context.Context, ...outbox.Message) error { return nil }

func newTestApp(t *testing.T) (*App, context.CancelFunc) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	rt := apphost.NewForTest(log)
	cfg := &config.Config{JwtSecret: "websocket-test-secret"}
	server := ws.NewServer(rt.Ctx, token.NewTokenProvider(cfg), log, accessCheckerStub{}, accessCheckerStub{}, time.Second, enqueuerStub{})
	return &App{rt: rt, Ws: server}, func() { rt.Close() }
}

func TestRouterDeliversRealtimeMessagesFromIndependentService(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()
	httpServer := httptest.NewServer(app.Router())
	defer httpServer.Close()

	userID := uuid.New()
	jwtProvider := token.NewTokenProvider(&config.Config{JwtSecret: "websocket-test-secret"})
	accessToken, err := jwtProvider.Generate(userID.String(), time.Now().Add(time.Minute), nil)
	require.NoError(t, err)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()
	connection, _, err := websocket.Dial(dialCtx, "ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws?jwt="+accessToken, nil)
	require.NoError(t, err)
	defer connection.Close(websocket.StatusNormalClosure, "test complete")

	require.NoError(t, wsjson.Write(dialCtx, connection, ws.WebsocketMessage{
		Type: ws.WebsocketMessageTypeConnectUserToRoom,
		Data: ws.ConnectUserToRoomData{RoomId: userID, Type: ws.WsRoomTypeUser},
	}))

	var connected ws.WebsocketMessage
	require.NoError(t, wsjson.Read(dialCtx, connection, &connected))
	require.Equal(t, ws.WebsocketMessageTypeUserConnected, connected.Type)

	notification := &domain.Notification{UserId: userID}
	require.NoError(t, app.Ws.SendNotification(context.Background(), notification))

	var delivered ws.WebsocketMessage
	require.NoError(t, wsjson.Read(dialCtx, connection, &delivered))
	require.Equal(t, ws.WebsocketMessageTypeNotificationCreated, delivered.Type)
	require.Equal(t, userID, delivered.RoomId)
}

func TestWebsocketListenerFailureDoesNotStopAPIListener(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	app, _ := newTestApp(t)
	websocketServer := httptest.NewServer(app.Router())
	websocketServer.Close()
	app.Close()

	response, err := http.Get(apiServer.URL)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
}
