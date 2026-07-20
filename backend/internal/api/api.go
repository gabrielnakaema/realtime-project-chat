package api

import (
	"errors"
	"net"
	"sync"

	"github.com/gabrielnakaema/project-chat/internal/apphost"
	"github.com/gabrielnakaema/project-chat/internal/auth"
	"github.com/gabrielnakaema/project-chat/internal/handlers"
	"github.com/gabrielnakaema/project-chat/internal/project"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/gabrielnakaema/project-chat/internal/repository"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/subscriber"
	"github.com/gabrielnakaema/project-chat/internal/token"
	"google.golang.org/grpc"
)

type Api struct {
	rt           *apphost.Runtime
	handlers     *Handlers
	grpcServer   *grpc.Server
	grpcStopOnce sync.Once
}

type Handlers struct {
	AuthMiddleware *auth.Middleware
	MCPAPIKey      *handlers.MCPAPIKeyHandler
	Project        *handlers.ProjectHandler
	User           *handlers.UserHandler
}

func NewApi() (*Api, error) {
	rt, err := apphost.New("api", "", "")
	if err != nil {
		return nil, err
	}

	jwtProvider := token.NewTokenProvider(rt.Config)
	authMiddleware := auth.NewMiddleware(jwtProvider)

	projectRepo := repository.NewProjectRepository(rt.Pool)
	userRepo := repository.NewUserRepository(rt.Pool)
	activityRepo := repository.NewProjectActivityRepository(rt.Pool)
	mcpAPIKeyRepo := repository.NewMCPAPIKeyRepository(rt.Pool)

	projectService := service.NewProjectService(projectRepo, userRepo, activityRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	grpcServer := NewInternalGRPCServer(project.NewServer(projectService))

	activitySub, err := subscriber.NewProjectActivitySubscriber(rt.Ctx, rt.Config, rt.Logger, activityRepo, projectRepo)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(activitySub)

	userService := service.NewUserService(jwtProvider, userRepo)
	userHandler := handlers.NewUserHandler(userService, rt.Config)
	mcpAPIKeyService := service.NewMCPAPIKeyService(mcpAPIKeyRepo)
	mcpAPIKeyHandler := handlers.NewMCPAPIKeyHandler(mcpAPIKeyService)

	api := Api{
		rt:         rt,
		grpcServer: grpcServer,
		handlers: &Handlers{
			AuthMiddleware: authMiddleware,
			MCPAPIKey:      mcpAPIKeyHandler,
			Project:        projectHandler,
			User:           userHandler,
		},
	}

	return &api, nil
}

func (a *Api) Close() {
	a.stopGRPC()
	a.rt.Close()
}

func NewInternalGRPCServer(projectServer projectv1.ProjectServiceServer) *grpc.Server {
	server := grpc.NewServer()
	projectv1.RegisterProjectServiceServer(server, projectServer)
	return server
}

func (a *Api) stopGRPC() {
	a.grpcStopOnce.Do(a.grpcServer.GracefulStop)
}

func (a *Api) Serve() error {
	listener, err := net.Listen("tcp", a.rt.Config.InternalGRPCListenAddress)
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
