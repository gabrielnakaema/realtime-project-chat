package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                          string
	JwtSecret                     string
	DSN                           string
	PubsubBrokers                 []string
	Environment                   string
	CORSOrigins                   []string
	InternalGRPCListenAddress     string
	AuthorizationGRPCTarget       string
	ChatInternalGRPCListenAddress string
	ChatAuthorizationGRPCTarget   string
	RoomAuthorizationTimeout      time.Duration
	OutboxPollInterval            time.Duration
	OutboxBatchSize               int32
}

func New() (*Config, error) {
	godotenv.Load()

	env := getEnv("ENV", "development")

	port := getEnv("API_PORT", "3333")

	authorizationTimeout, err := time.ParseDuration(getEnv("ROOM_AUTHORIZATION_TIMEOUT", "2s"))
	if err != nil {
		return nil, fmt.Errorf("parse ROOM_AUTHORIZATION_TIMEOUT: %w", err)
	}
	if authorizationTimeout <= 0 {
		return nil, fmt.Errorf("ROOM_AUTHORIZATION_TIMEOUT must be greater than zero")
	}

	outboxPollInterval, err := time.ParseDuration(getEnv("OUTBOX_POLL_INTERVAL", "500ms"))
	if err != nil {
		return nil, fmt.Errorf("parse OUTBOX_POLL_INTERVAL: %w", err)
	}
	if outboxPollInterval <= 0 {
		return nil, fmt.Errorf("OUTBOX_POLL_INTERVAL must be greater than zero")
	}

	outboxBatchSize, err := strconv.Atoi(getEnv("OUTBOX_BATCH_SIZE", "100"))
	if err != nil {
		return nil, fmt.Errorf("parse OUTBOX_BATCH_SIZE: %w", err)
	}
	if outboxBatchSize <= 0 {
		return nil, fmt.Errorf("OUTBOX_BATCH_SIZE must be greater than zero")
	}

	config := Config{
		Port:                          port,
		DSN:                           getEnv("DB_DSN", ""),
		PubsubBrokers:                 strings.Split(getEnv("PUBSUB_BROKERS", ""), ","),
		JwtSecret:                     getEnv("JWT_SECRET", "SECRET"),
		Environment:                   env,
		CORSOrigins:                   strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
		InternalGRPCListenAddress:     getEnv("INTERNAL_GRPC_LISTEN_ADDRESS", "127.0.0.1:3334"),
		AuthorizationGRPCTarget:       getEnv("AUTHORIZATION_GRPC_TARGET", "127.0.0.1:3334"),
		ChatInternalGRPCListenAddress: getEnv("CHAT_INTERNAL_GRPC_LISTEN_ADDRESS", "127.0.0.1:3338"),
		ChatAuthorizationGRPCTarget:   getEnv("CHAT_AUTHORIZATION_GRPC_TARGET", "127.0.0.1:3338"),
		RoomAuthorizationTimeout:      authorizationTimeout,
		OutboxPollInterval:            outboxPollInterval,
		OutboxBatchSize:               int32(outboxBatchSize),
	}

	return &config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
