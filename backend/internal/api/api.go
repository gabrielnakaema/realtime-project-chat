package api

import (
	"github.com/gabrielnakaema/project-chat/internal/apphost"
	"github.com/gabrielnakaema/project-chat/internal/auth"
	"github.com/gabrielnakaema/project-chat/internal/handlers"
	"github.com/gabrielnakaema/project-chat/internal/mcp"
	"github.com/gabrielnakaema/project-chat/internal/repository"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/subscriber"
	"github.com/gabrielnakaema/project-chat/internal/token"
)

type Api struct {
	rt       *apphost.Runtime
	handlers *Handlers
}

type Handlers struct {
	AuthMiddleware *auth.Middleware
	Chat           *handlers.ChatHandler
	MCPAPIKey      *handlers.MCPAPIKeyHandler
	MCP            *mcp.Handler
	Project        *handlers.ProjectHandler
	Task           *handlers.TaskHandler
	TaskComment    *handlers.TaskCommentHandler
	User           *handlers.UserHandler
}

func NewApi() (*Api, error) {
	rt, err := apphost.New("api", "", "")
	if err != nil {
		return nil, err
	}

	jwtProvider := token.NewTokenProvider(rt.Config)
	authMiddleware := auth.NewMiddleware(jwtProvider)

	chatRepo := repository.NewChatRepository(rt.Pool)
	projectRepo := repository.NewProjectRepository(rt.Pool)
	taskRepo := repository.NewTaskRepository(rt.Pool)
	taskCommentRepo := repository.NewTaskCommentRepository(rt.Pool)
	userRepo := repository.NewUserRepository(rt.Pool)
	activityRepo := repository.NewProjectActivityRepository(rt.Pool)
	mcpAPIKeyRepo := repository.NewMCPAPIKeyRepository(rt.Pool)

	projectService := service.NewProjectService(projectRepo, userRepo, rt.Publisher, activityRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	chatService := service.NewChatService(chatRepo, userRepo, rt.Publisher)

	chatSub, err := subscriber.NewChatSubscriber(rt.Ctx, rt.Config, rt.Logger, chatService)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(chatSub)

	activitySub, err := subscriber.NewProjectActivitySubscriber(rt.Ctx, rt.Config, rt.Logger, activityRepo, projectRepo)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(activitySub)

	taskUpdateSub, err := subscriber.NewTaskUpdateSubscriber(rt.Ctx, rt.Config, rt.Logger, taskRepo)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(taskUpdateSub)

	chatHandler := handlers.NewChatHandler(chatService)

	userService := service.NewUserService(jwtProvider, userRepo)
	userHandler := handlers.NewUserHandler(userService, rt.Config)
	mcpAPIKeyService := service.NewMCPAPIKeyService(mcpAPIKeyRepo)
	mcpAPIKeyHandler := handlers.NewMCPAPIKeyHandler(mcpAPIKeyService)

	taskService := service.NewTaskService(taskRepo, projectRepo, userRepo, rt.Publisher)
	taskHandler := handlers.NewTaskHandler(taskService)
	taskCommentService := service.NewTaskCommentService(taskCommentRepo, taskRepo, projectRepo, userRepo, rt.Publisher)
	taskCommentHandler := handlers.NewTaskCommentHandler(taskCommentService)
	mcpHandler := mcp.NewHandler(mcpAPIKeyService, projectService, taskService, taskCommentService)

	api := Api{
		rt: rt,
		handlers: &Handlers{
			AuthMiddleware: authMiddleware,
			Chat:           chatHandler,
			MCPAPIKey:      mcpAPIKeyHandler,
			MCP:            mcpHandler,
			Project:        projectHandler,
			Task:           taskHandler,
			TaskComment:    taskCommentHandler,
			User:           userHandler,
		},
	}

	return &api, nil
}

func (a *Api) Close() {
	a.rt.Close()
}

func (a *Api) Serve() error {
	return a.rt.Serve(a.Router())
}
