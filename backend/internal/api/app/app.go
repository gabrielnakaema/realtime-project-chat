package app

import (
	"errors"
	"net"
	"sync"

	"github.com/gabrielnakaema/project-chat/internal/apikey"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	"github.com/gabrielnakaema/project-chat/internal/project"
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
	MCPAPIKey      *apikey.MCPAPIKeyHandler
	Project        *project.ProjectHandler
	User           *user.UserHandler
}

func New(rt *apphost.Runtime, handlers *Handlers, grpcServer *grpc.Server) *App {
	return &App{
		rt:         rt,
		grpcServer: grpcServer,
		handlers:   handlers,
	}
}

func (a *App) Close() {
	a.stopGRPC()
	a.rt.Close()
}

func (a *App) stopGRPC() {
	a.grpcStopOnce.Do(a.grpcServer.GracefulStop)
}

func (a *App) Serve() error {
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
