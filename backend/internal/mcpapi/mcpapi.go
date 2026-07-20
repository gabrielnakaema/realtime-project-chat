package mcpapi

import (
	"github.com/gabrielnakaema/project-chat/internal/apphost"
	"github.com/gabrielnakaema/project-chat/internal/mcp"
	"github.com/gabrielnakaema/project-chat/internal/repository"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/tasks"
	tasksv1 "github.com/gabrielnakaema/project-chat/internal/tasks/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	rt      *apphost.Runtime
	handler *mcp.Handler
}

func New() (*App, error) {
	rt, err := apphost.New("mcp-service", "MCP_SERVICE_PORT", "3341")
	if err != nil {
		return nil, err
	}

	mcpAPIKeyRepo := repository.NewMCPAPIKeyRepository(rt.Pool)
	projectRepo := repository.NewProjectRepository(rt.Pool)
	userRepo := repository.NewUserRepository(rt.Pool)
	activityRepo := repository.NewProjectActivityRepository(rt.Pool)

	mcpAPIKeyService := service.NewMCPAPIKeyService(mcpAPIKeyRepo)
	projectService := service.NewProjectService(projectRepo, userRepo, activityRepo)

	tasksGRPCConnection, err := grpc.NewClient(
		rt.Config.TasksAuthorizationGRPCTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(tasksGRPCConnection)

	tasksClient := tasksv1.NewTaskServiceClient(tasksGRPCConnection)
	taskServiceClient := tasks.NewTaskServiceClient(tasksClient)
	taskCommentServiceClient := tasks.NewTaskCommentServiceClient(tasksClient)

	mcpHandler := mcp.NewHandler(mcpAPIKeyService, projectService, taskServiceClient, taskCommentServiceClient)

	app := App{
		rt:      rt,
		handler: mcpHandler,
	}

	return &app, nil
}

func (a *App) Close() {
	a.rt.Close()
}

func (a *App) Serve() error {
	return a.rt.Serve(a.Router())
}
