package notificationapi

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

	"github.com/gabrielnakaema/project-chat/internal/auth"
	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/db"
	"github.com/gabrielnakaema/project-chat/internal/logger"
	"github.com/gabrielnakaema/project-chat/internal/notification"
	"github.com/gabrielnakaema/project-chat/internal/publisher"
	"github.com/gabrielnakaema/project-chat/internal/token"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	config      *config.Config
	pool        *pgxpool.Pool
	handlers    *Handlers
	logger      *slog.Logger
	Publisher   *publisher.Publisher
	subscribers []io.Closer
	cancel      context.CancelFunc
}

type Handlers struct {
	AuthMiddleware *auth.Middleware
	Notification   *notification.Handler
}

func New() (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	if port := os.Getenv("NOTIFICATION_SERVICE_PORT"); port != "" {
		cfg.Port = port
	} else {
		cfg.Port = "3335"
	}

	log := logger.Init(cfg)

	pool, err := db.NewPool(cfg)
	if err != nil {
		return nil, err
	}

	pub, err := publisher.NewPublisher(cfg, log)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	jwtProvider := token.NewTokenProvider(cfg)
	authMiddleware := auth.NewMiddleware(jwtProvider)

	notificationRepo := notification.NewRepository(pool)
	notificationService := notification.NewService(notificationRepo)
	notificationHandler := notification.NewHandler(notificationService)

	var subscribers []io.Closer

	notificationSub, err := notification.NewEventSubscriber(ctx, cfg, log, notificationRepo, notification.NewKafkaNotifier(pub))
	if err != nil {
		cancel()
		return nil, err
	}
	subscribers = append(subscribers, notificationSub)

	h := &Handlers{
		AuthMiddleware: authMiddleware,
		Notification:   notificationHandler,
	}

	app := App{
		handlers:    h,
		config:      cfg,
		pool:        pool,
		logger:      log,
		Publisher:   pub,
		subscribers: subscribers,
		cancel:      cancel,
	}

	return &app, nil
}

func (a *App) Close() {
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

func (a *App) Serve() error {
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

	a.logger.Info("starting notification-service", "addr", addr, "environment", a.config.Environment)

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	a.logger.Info("stopped notification-service", "addr", addr, "environment", a.config.Environment)

	return nil
}
