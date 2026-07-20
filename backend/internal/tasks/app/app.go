package app

import (
	"errors"
	"net"
	"sync"

	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	"github.com/gabrielnakaema/project-chat/internal/platform/postgres"
	"github.com/gabrielnakaema/project-chat/internal/platform/token"
	"github.com/gabrielnakaema/project-chat/internal/project"
	"github.com/gabrielnakaema/project-chat/internal/tasks"
	tasksv1 "github.com/gabrielnakaema/project-chat/internal/tasks/v1"
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
	Task           *tasks.TaskHandler
	TaskComment    *tasks.TaskCommentHandler
}

func New() (*App, error) {
	rt, err := apphost.New("tasks-service", "TASKS_SERVICE_PORT", "3339")
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

	taskRepo := tasks.NewTaskRepository(pool)
	taskCommentRepo := tasks.NewTaskCommentRepository(pool)
	projectRepo := project.NewProjectRepository(pool)
	userRepo := user.NewUserRepository(pool)

	taskService := tasks.NewTaskService(taskRepo, projectRepo, userRepo)
	taskHandler := tasks.NewTaskHandler(taskService)
	taskCommentService := tasks.NewTaskCommentService(taskCommentRepo, taskRepo, projectRepo, userRepo)
	taskCommentHandler := tasks.NewTaskCommentHandler(taskCommentService)

	taskUpdateSub, err := tasks.NewTaskUpdateSubscriber(rt.Ctx, rt.Config, rt.Logger, taskRepo)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(taskUpdateSub)

	memberCleanupSub, err := tasks.NewMemberCleanupSubscriber(rt.Ctx, rt.Config, rt.Logger, taskRepo)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(memberCleanupSub)

	grpcServer := NewInternalGRPCServer(tasks.NewGRPCServer(taskService, taskCommentService))

	app := App{
		rt:         rt,
		grpcServer: grpcServer,
		handlers: &Handlers{
			AuthMiddleware: authMiddleware,
			Task:           taskHandler,
			TaskComment:    taskCommentHandler,
		},
	}

	return &app, nil
}

func (a *App) Close() {
	a.stopGRPC()
	a.rt.Close()
}

func NewInternalGRPCServer(taskServer tasksv1.TaskServiceServer) *grpc.Server {
	server := grpc.NewServer()
	tasksv1.RegisterTaskServiceServer(server, taskServer)
	return server
}

func (a *App) stopGRPC() {
	a.grpcStopOnce.Do(a.grpcServer.GracefulStop)
}

func (a *App) Serve() error {
	listener, err := net.Listen("tcp", a.rt.Config.TasksInternalGRPCListenAddress)
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
