package app

import (
	"errors"
	"net"
	"sync"

	"github.com/gabrielnakaema/project-chat/internal/chat"
	chatv1 "github.com/gabrielnakaema/project-chat/internal/chat/v1"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	"github.com/gabrielnakaema/project-chat/internal/platform/postgres"
	"github.com/gabrielnakaema/project-chat/internal/platform/token"
	"github.com/gabrielnakaema/project-chat/internal/user"
	"google.golang.org/grpc"
)

type App struct {
	rt           *apphost.Runtime
	handlers     *Handlers
	grpcServer   *grpc.Server
	grpcStopOnce sync.Once
}

type Handlers struct {
	AuthMiddleware *auth.Middleware
	Chat           *chat.Handler
}

func New() (*App, error) {
	rt, err := apphost.New("chat-service", "CHAT_SERVICE_PORT", "3337")
	if err != nil {
		return nil, err
	}
	pool, err := postgres.NewPool(rt.Config)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.TrackFunc(func() error {
		pool.Close()
		return nil
	})

	jwtProvider := token.NewTokenProvider(rt.Config)
	authMiddleware := auth.NewMiddleware(jwtProvider)

	chatRepo := chat.NewRepository(pool)
	userRepo := user.NewUserRepository(pool)
	chatService := chat.NewService(chatRepo, userRepo)
	chatHandler := chat.NewHandler(chatService)

	chatSub, err := chat.NewEventSubscriber(rt.Ctx, rt.Config, rt.Logger, chatService)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(chatSub)

	grpcServer := NewInternalGRPCServer(chat.NewServer(chatService))

	app := App{
		rt:         rt,
		grpcServer: grpcServer,
		handlers: &Handlers{
			AuthMiddleware: authMiddleware,
			Chat:           chatHandler,
		},
	}

	return &app, nil
}

func (a *App) Close() {
	a.stopGRPC()
	a.rt.Close()
}

func NewInternalGRPCServer(chatServer chatv1.ChatServiceServer) *grpc.Server {
	server := grpc.NewServer()
	chatv1.RegisterChatServiceServer(server, chatServer)
	return server
}

func (a *App) stopGRPC() {
	a.grpcStopOnce.Do(a.grpcServer.GracefulStop)
}

func (a *App) Serve() error {
	listener, err := net.Listen("tcp", a.rt.Config.ChatInternalGRPCListenAddress)
	if err != nil {
		return err
	}

	grpcError := make(chan error, 1)
	go func() {
		a.rt.Logger.Info("starting internal gRPC server", "service", a.rt.Name, "addr", listener.Addr().String())
		grpcError <- a.grpcServer.Serve(listener)
	}()

	httpError := a.rt.Serve(a.Router())
	a.stopGRPC()
	serveError := <-grpcError
	if httpError != nil {
		return httpError
	}
	if serveError != nil && !errors.Is(serveError, grpc.ErrServerStopped) {
		return serveError
	}
	return nil
}
