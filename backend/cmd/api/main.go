package main

import (
	"log"

	"github.com/gabrielnakaema/project-chat/internal/api/app"
	"github.com/gabrielnakaema/project-chat/internal/apikey"
	apikeyv1 "github.com/gabrielnakaema/project-chat/internal/apikey/v1"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	"github.com/gabrielnakaema/project-chat/internal/platform/postgres"
	"github.com/gabrielnakaema/project-chat/internal/platform/token"
	"github.com/gabrielnakaema/project-chat/internal/project"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/gabrielnakaema/project-chat/internal/user"
	"google.golang.org/grpc"
)

func main() {
	a, err := newAPI()
	if err != nil {
		log.Fatal("error while starting api", "error", err.Error())
		return
	}

	defer a.Close()

	err = a.Serve()
	if err != nil {
		log.Fatal("received error from api serve", "error", err.Error())
		return
	}
}

func newAPI() (*app.App, error) {
	rt, err := apphost.New("api", "", "")
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

	userService := user.NewUserService(jwtProvider, userRepo)

	grpcServer := grpc.NewServer()
	projectv1.RegisterProjectServiceServer(grpcServer, project.NewServer(projectService))
	apikeyv1.RegisterAPIKeyServiceServer(grpcServer, apikey.NewServer(mcpAPIKeyService))

	return app.New(rt, &app.Handlers{
		AuthMiddleware: authMiddleware,
		MCPAPIKey:      apikey.NewMCPAPIKeyHandler(mcpAPIKeyService),
		Project:        project.NewProjectHandler(projectService),
		User:           user.NewUserHandler(userService, rt.Config),
	}, grpcServer), nil
}
