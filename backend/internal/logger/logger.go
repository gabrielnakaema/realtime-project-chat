package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/gabrielnakaema/project-chat/internal/config"
)

func Init(config *config.Config) *slog.Logger {
	environment := config.Environment
	if config.Environment == "" {
		environment = "development"
	}

	addSource := false
	logLevel := slog.LevelInfo

	if environment == "development" {
		addSource = false
		logLevel = slog.LevelDebug
	}

	handlerOptions := &slog.HandlerOptions{AddSource: addSource, Level: logLevel}

	var logger *slog.Logger

	if environment == "test" {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	}

	slog.SetDefault(logger)

	return logger
}

var ServerErrorMsg string = "internal_server_error"

type loggerCtxKey string

const LoggerCtxKey loggerCtxKey = "logger"

func FromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(LoggerCtxKey).(*slog.Logger)
	if !ok || logger == nil {
		return slog.Default()
	}
	return logger
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerCtxKey, logger)
}
