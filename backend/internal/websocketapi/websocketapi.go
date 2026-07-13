package websocketapi

import (
	"github.com/gabrielnakaema/project-chat/internal/apphost"
	"github.com/gabrielnakaema/project-chat/internal/notification"
	"github.com/gabrielnakaema/project-chat/internal/repository"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/subscriber"
	"github.com/gabrielnakaema/project-chat/internal/token"
	"github.com/gabrielnakaema/project-chat/internal/ws"
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
	chatRepo := repository.NewChatRepository(rt.Pool)
	projectRepo := repository.NewProjectRepository(rt.Pool)
	userRepo := repository.NewUserRepository(rt.Pool)
	activityRepo := repository.NewProjectActivityRepository(rt.Pool)
	chatService := service.NewChatService(chatRepo, userRepo, rt.Publisher)
	projectService := service.NewProjectService(projectRepo, userRepo, rt.Publisher, activityRepo)
	server := ws.NewServer(rt.Ctx, jwtProvider, rt.Logger, chatService, projectService, rt.Publisher)

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
