package api

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	config      *config.Config
	pool        *pgxpool.Pool
	handlers    *Handlers
	logger      *slog.Logger
	Publisher   *publisher.Publisher
	Ws          *ws.Server
	subscribers []io.Closer
	cancel      context.CancelFunc
}

type Handlers struct {
	AuthMiddleware *handlers.AuthMiddleware
	Chat           *handlers.ChatHandler
	Project        *handlers.ProjectHandler
	Task           *handlers.TaskHandler
	TaskComment    *handlers.TaskCommentHandler
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

	ctx, cancel := context.WithCancel(context.Background())

	jwtProvider := token.NewTokenProvider(config)
	authMiddleware := handlers.NewAuthMiddleware(jwtProvider)

	chatRepo := repository.NewChatRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)
	taskCommentRepo := repository.NewTaskCommentRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	activityRepo := repository.NewProjectActivityRepository(pool)

	projectService := service.NewProjectService(projectRepo, userRepo, pub, activityRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	chatService := service.NewChatService(chatRepo, userRepo, pub)

	ws := ws.NewServer(ctx, jwtProvider, logger, chatService, projectService, pub)

	var subscribers []io.Closer

	chatSub, err := subscriber.NewChatSubscriber(ctx, config, logger, chatService, ws)
	if err != nil {
		cancel()
		return nil, err
	}
	subscribers = append(subscribers, chatSub)

	taskSub, err := subscriber.NewTaskSubscriber(ctx, config, logger, ws)
	if err != nil {
		cancel()
		return nil, err
	}
	subscribers = append(subscribers, taskSub)

	activitySub, err := subscriber.NewProjectActivitySubscriber(ctx, config, logger, activityRepo, projectRepo)
	if err != nil {
		cancel()
		return nil, err
	}
	subscribers = append(subscribers, activitySub)

	taskUpdateSub, err := subscriber.NewTaskUpdateSubscriber(ctx, config, logger, taskRepo)
	if err != nil {
		cancel()
		return nil, err
	}
	subscribers = append(subscribers, taskUpdateSub)

	chatHandler := handlers.NewChatHandler(chatService)

	userService := service.NewUserService(jwtProvider, userRepo)
	userHandler := handlers.NewUserHandler(userService, config)

	taskService := service.NewTaskService(taskRepo, projectRepo, userRepo, pub)
	taskHandler := handlers.NewTaskHandler(taskService)
	taskCommentService := service.NewTaskCommentService(taskCommentRepo, taskRepo, projectRepo, userRepo, pub)
	taskCommentHandler := handlers.NewTaskCommentHandler(taskCommentService)

	handlers := Handlers{
		AuthMiddleware: authMiddleware,
		Chat:           chatHandler,
		Project:        projectHandler,
		Task:           taskHandler,
		TaskComment:    taskCommentHandler,
		User:           userHandler,
	}

	api := Api{
		handlers:    &handlers,
		config:      config,
		pool:        pool,
		logger:      logger,
		Ws:          ws,
		Publisher:   pub,
		subscribers: subscribers,
		cancel:      cancel,
	}

	return &api, nil
}

func (a *Api) Close() {
	a.cancel()

	for _, s := range a.subscribers {
		if err := s.Close(); err != nil {
			a.logger.Error("error closing subscriber", "error", err)
		}
	}

	if err := a.Publisher.Close(); err != nil {
		a.logger.Error("error closing publisher", "error", err)
	}

	a.pool.Close()

	a.logger.Info("all resources closed")
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
