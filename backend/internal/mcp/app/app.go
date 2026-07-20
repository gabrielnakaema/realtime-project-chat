package app

import (
	apikeyv1 "github.com/gabrielnakaema/project-chat/internal/apikey/v1"
	"github.com/gabrielnakaema/project-chat/internal/mcp"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
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

	apiGRPCConnection, err := grpc.NewClient(
		rt.Config.AuthorizationGRPCTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(apiGRPCConnection)

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
	mcpHandler := mcp.NewHandler(
		newAPIKeyService(apikeyv1.NewAPIKeyServiceClient(apiGRPCConnection)),
		newProjectService(projectv1.NewProjectServiceClient(apiGRPCConnection)),
		newTaskService(tasksClient),
		newTaskCommentService(tasksClient),
	)

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
