package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/db"
	"github.com/gabrielnakaema/project-chat/internal/handlers"
	"github.com/gabrielnakaema/project-chat/internal/logger"
	"github.com/gabrielnakaema/project-chat/internal/publisher"
	"github.com/gabrielnakaema/project-chat/internal/repository"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/subscriber"
	"github.com/gabrielnakaema/project-chat/internal/token"
	"github.com/gabrielnakaema/project-chat/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Api struct {
	config    *config.Config
	pool      *pgxpool.Pool
	handlers  *Handlers
	logger    *slog.Logger
	Publisher *publisher.Publisher
	Ws        *ws.Server
}

type Handlers struct {
	AuthMiddleware *handlers.AuthMiddleware
	Chat           *handlers.ChatHandler
	Project        *handlers.ProjectHandler
	Task           *handlers.TaskHandler
	User           *handlers.UserHandler
}

func NewApi() (*Api, error) {
	config, err := config.New()
	if err != nil {
		return nil, err
	}

	logger := logger.Init(config)

	pool, err := db.NewPool(config)
	if err != nil {
		return nil, err
	}

	pub, err := publisher.NewPublisher(config, logger)
	if err != nil {
		return nil, err
	}

	jwtProvider := token.NewTokenProvider(config)
	authMiddleware := handlers.NewAuthMiddleware(jwtProvider)

	chatRepo := repository.NewChatRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	projectService := service.NewProjectService(projectRepo, userRepo, pub)
	projectHandler := handlers.NewProjectHandler(projectService)

	chatService := service.NewChatService(chatRepo, userRepo, pub)

	ws := ws.NewServer(jwtProvider, logger, chatService, projectService, pub)

	_, err = subscriber.NewChatSubscriber(config, logger, chatService, ws)
	if err != nil {
		return nil, err
	}

	_, err = subscriber.NewTaskSubscriber(config, logger, ws)
	if err != nil {
		return nil, err
	}

	chatHandler := handlers.NewChatHandler(chatService)

	userService := service.NewUserService(jwtProvider, userRepo)
	userHandler := handlers.NewUserHandler(userService)

	taskService := service.NewTaskService(taskRepo, projectRepo, userRepo, pub)
	taskHandler := handlers.NewTaskHandler(taskService)

	handlers := Handlers{
		AuthMiddleware: authMiddleware,
		Chat:           chatHandler,
		Project:        projectHandler,
		Task:           taskHandler,
		User:           userHandler,
	}

	api := Api{
		handlers:  &handlers,
		config:    config,
		pool:      pool,
		logger:    logger,
		Ws:        ws,
		Publisher: pub,
	}

	return &api, nil
}

func (a *Api) Serve() error {
	addr := fmt.Sprintf(":%s", a.config.Port)

	server := &http.Server{
		Addr:         addr,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      a.Router(),
		ErrorLog:     slog.NewLogLogger(a.logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		a.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownError <- server.Shutdown(ctx)
	}()

	a.logger.Info("starting server", "addr", addr, "environment", a.config.Environment)

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	a.logger.Info("stopped server", "addr", addr, "environment", a.config.Environment)

	return nil
}
