package apphost

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	"github.com/gabrielnakaema/project-chat/internal/platform/config"
	"github.com/gabrielnakaema/project-chat/internal/platform/logger"
	"github.com/gabrielnakaema/project-chat/internal/platform/postgres"
	"github.com/gabrielnakaema/project-chat/internal/platform/token"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application interface {
	Serve() error
	Close()
}

type AuthDependencies struct {
	TokenProvider *token.TokenProvider
	Middleware    *auth.Middleware
}

type Runtime struct {
	Name   string
	Config *config.Config
	Logger *slog.Logger
	Ctx    context.Context

	cancel      context.CancelFunc
	subscribers []io.Closer
	closeOnce   sync.Once
}

type closerFunc func() error

func (f closerFunc) Close() error {
	return f()
}

func Run[T Application](serviceName string, factory func() (T, error)) error {
	application, err := factory()
	if err != nil {
		return fmt.Errorf("error while starting %s: %w", serviceName, err)
	}
	defer application.Close()

	if err := application.Serve(); err != nil {
		return fmt.Errorf("received error from %s serve: %w", serviceName, err)
	}

	return nil
}

func New(name, portEnvVar, defaultPort string) (*Runtime, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	if portEnvVar != "" {
		if port := os.Getenv(portEnvVar); port != "" {
			cfg.Port = port
		} else if defaultPort != "" {
			cfg.Port = defaultPort
		}
	}

	log := logger.Init(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	return &Runtime{
		Name:   name,
		Config: cfg,
		Logger: log,
		Ctx:    ctx,
		cancel: cancel,
	}, nil
}

func NewWithPostgres(name, portEnvVar, defaultPort string) (*Runtime, *pgxpool.Pool, error) {
	rt, err := New(name, portEnvVar, defaultPort)
	if err != nil {
		return nil, nil, err
	}

	pool, err := postgres.NewPool(rt.Config)
	if err != nil {
		rt.Close()
		return nil, nil, err
	}
	rt.TrackFunc(func() error {
		pool.Close()
		return nil
	})

	return rt, pool, nil
}

func (r *Runtime) NewAuth() *AuthDependencies {
	tokenProvider := token.NewTokenProvider(r.Config)
	return &AuthDependencies{
		TokenProvider: tokenProvider,
		Middleware:    auth.NewMiddleware(tokenProvider),
	}
}

func (r *Runtime) Track(c io.Closer) {
	r.subscribers = append(r.subscribers, c)
}

func (r *Runtime) TrackFunc(close func() error) {
	r.Track(closerFunc(close))
}

func (r *Runtime) Close() {
	r.closeOnce.Do(func() {
		if r.cancel != nil {
			r.cancel()
		}

		for i := len(r.subscribers) - 1; i >= 0; i-- {
			if err := r.subscribers[i].Close(); err != nil {
				r.Logger.Error("error closing resource", "service", r.Name, "error", err)
			}
		}

		r.Logger.Info("all resources closed", "service", r.Name)
	})
}

func NewForTest(log *slog.Logger) *Runtime {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runtime{
		Name:   "test",
		Config: &config.Config{},
		Logger: log,
		Ctx:    ctx,
		cancel: cancel,
	}
}

func (r *Runtime) Serve(handler http.Handler) error {
	addr := fmt.Sprintf(":%s", r.Config.Port)

	server := &http.Server{
		Addr:         addr,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler,
		ErrorLog:     slog.NewLogLogger(r.Logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error, 1)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(quit)

		s := <-quit
		r.Logger.Info("shutting down", "service", r.Name, "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownError <- server.Shutdown(ctx)
	}()

	r.Logger.Info("starting", "service", r.Name, "addr", addr, "environment", r.Config.Environment)

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	r.Logger.Info("stopped", "service", r.Name, "addr", addr, "environment", r.Config.Environment)

	return nil
}
