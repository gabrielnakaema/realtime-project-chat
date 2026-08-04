package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/api/app"
	"github.com/gabrielnakaema/project-chat/internal/apikey"
	apikeyv1 "github.com/gabrielnakaema/project-chat/internal/apikey/v1"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	"github.com/gabrielnakaema/project-chat/internal/project"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/gabrielnakaema/project-chat/internal/user"
	"google.golang.org/grpc"
)

func main() {
	if err := apphost.Run("api", newAPI); err != nil {
		log.Fatal(err)
	}
}

func newAPI() (*app.App, error) {
	rt, pool, err := apphost.NewWithPostgres("api", "", "")
	if err != nil {
		return nil, err
	}
	authDependencies := rt.NewAuth()

	projectRepo := project.NewProjectRepository(pool)
	userRepo := user.NewUserRepository(pool)
	activityRepo := project.NewProjectActivityRepository(pool)
	mcpAPIKeyRepo := apikey.NewMCPAPIKeyRepository(pool)

	projectService := project.NewProjectService(projectRepo, userRepo, activityRepo)
	mcpAPIKeyService := apikey.NewMCPAPIKeyService(mcpAPIKeyRepo)

	activitySub, err := project.NewProjectActivitySubscriber(rt.Ctx, rt.Config, rt.Logger, activityRepo, projectRepo)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(activitySub)

	userService := user.NewUserService(authDependencies.TokenProvider, userRepo)

	grpcServer := grpc.NewServer()
	projectv1.RegisterProjectServiceServer(grpcServer, project.NewServer(projectService))
	apikeyv1.RegisterAPIKeyServiceServer(grpcServer, apikey.NewServer(mcpAPIKeyService))

	return app.New(rt, &app.Handlers{
		AuthMiddleware: authDependencies.Middleware,
		MCPAPIKey:      apikey.NewMCPAPIKeyHandler(mcpAPIKeyService),
		Project:        project.NewProjectHandler(projectService),
		User:           user.NewUserHandler(userService, rt.Config),
	}, grpcServer), nil
}
