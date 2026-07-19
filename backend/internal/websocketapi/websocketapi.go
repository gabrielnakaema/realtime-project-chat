package websocketapi

import (
	"github.com/gabrielnakaema/project-chat/internal/apphost"
	"github.com/gabrielnakaema/project-chat/internal/chat"
	chatv1 "github.com/gabrielnakaema/project-chat/internal/chat/v1"
	"github.com/gabrielnakaema/project-chat/internal/notification"
	"github.com/gabrielnakaema/project-chat/internal/outbox"
	"github.com/gabrielnakaema/project-chat/internal/project"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/gabrielnakaema/project-chat/internal/subscriber"
	"github.com/gabrielnakaema/project-chat/internal/token"
	"github.com/gabrielnakaema/project-chat/internal/ws"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	rt *apphost.Runtime
	Ws *ws.Server
}

func New() (*App, error) {
	rt, err := apphost.New("websocket-service", "WEBSOCKET_SERVICE_PORT", "3336")
	if err != nil {
		return nil, err
	}

	jwtProvider := token.NewTokenProvider(rt.Config)

	projectGRPCConnection, err := grpc.NewClient(
		rt.Config.AuthorizationGRPCTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(projectGRPCConnection)

	chatGRPCConnection, err := grpc.NewClient(
		rt.Config.ChatAuthorizationGRPCTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(chatGRPCConnection)

	chatAuthorizer := chat.NewClient(chatv1.NewChatServiceClient(chatGRPCConnection))
	projectAuthorizer := project.NewClient(projectv1.NewProjectServiceClient(projectGRPCConnection))
	server := ws.NewServer(rt.Ctx, jwtProvider, rt.Logger, chatAuthorizer, projectAuthorizer, rt.Config.RoomAuthorizationTimeout, outbox.NewPoolEnqueuer(rt.Pool))

	realtimeSub, err := subscriber.NewRealtimeSubscriber(rt.Ctx, rt.Config, rt.Logger, server)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(realtimeSub)

	taskSub, err := subscriber.NewTaskSubscriber(rt.Ctx, rt.Config, rt.Logger, server)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(taskSub)

	projectSub, err := subscriber.NewProjectSubscriber(rt.Ctx, rt.Config, rt.Logger, server)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(projectSub)

	notificationSub, err := notification.NewForwardSubscriber(rt.Ctx, rt.Config, rt.Logger, server)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(notificationSub)

	return &App{rt: rt, Ws: server}, nil
}

func (a *App) Close() {
	a.rt.Close()
}

func (a *App) Serve() error {
	return a.rt.Serve(a.Router())
}
