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

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/gabrielnakaema/project-chat/internal/db"
	"github.com/gabrielnakaema/project-chat/internal/logger"
	"github.com/gabrielnakaema/project-chat/internal/publisher"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Runtime struct {
	Name      string
	Config    *config.Config
	Pool      *pgxpool.Pool
	Publisher *publisher.Publisher
	Logger    *slog.Logger
	Ctx       context.Context

	cancel      context.CancelFunc
	subscribers []io.Closer
	closeOnce   sync.Once
}

func New(name, portEnvVar, defaultPort string) (*Runtime, error) {
	return newRuntime(name, portEnvVar, defaultPort, true)
}

func NewWithoutDB(name, portEnvVar, defaultPort string) (*Runtime, error) {
	return newRuntime(name, portEnvVar, defaultPort, false)
}

func newRuntime(name, portEnvVar, defaultPort string, withDB bool) (*Runtime, error) {
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

	var pool *pgxpool.Pool
	if withDB {
		pool, err = db.NewPool(cfg)
		if err != nil {
			return nil, err
		}
	}

	pub, err := publisher.NewPublisher(cfg, log)
	if err != nil {
		if pool != nil {
			pool.Close()
		}
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Runtime{
		Name:      name,
		Config:    cfg,
		Pool:      pool,
		Publisher: pub,
		Logger:    log,
		Ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (r *Runtime) Track(c io.Closer) {
	r.subscribers = append(r.subscribers, c)
}

func (r *Runtime) Close() {
	r.closeOnce.Do(func() {
		if r.cancel != nil {
			r.cancel()
		}

		for i := len(r.subscribers) - 1; i >= 0; i-- {
			if err := r.subscribers[i].Close(); err != nil {
				r.Logger.Error("error closing subscriber", "service", r.Name, "error", err)
			}
		}

		if r.Publisher != nil {
			if err := r.Publisher.Close(); err != nil {
				r.Logger.Error("error closing publisher", "service", r.Name, "error", err)
			}
		}

		if r.Pool != nil {
			r.Pool.Close()
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
