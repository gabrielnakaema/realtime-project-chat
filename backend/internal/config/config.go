package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                      string
	JwtSecret                 string
	DSN                       string
	PubsubBrokers             []string
	Environment               string
	CORSOrigins               []string
	InternalGRPCListenAddress string
	AuthorizationGRPCTarget   string
	RoomAuthorizationTimeout  time.Duration
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

	config := Config{
		Port:                      port,
		DSN:                       getEnv("DB_DSN", ""),
		PubsubBrokers:             strings.Split(getEnv("PUBSUB_BROKERS", ""), ","),
		JwtSecret:                 getEnv("JWT_SECRET", "SECRET"),
		Environment:               env,
		CORSOrigins:               strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
		InternalGRPCListenAddress: getEnv("INTERNAL_GRPC_LISTEN_ADDRESS", "127.0.0.1:3334"),
		AuthorizationGRPCTarget:   getEnv("AUTHORIZATION_GRPC_TARGET", "127.0.0.1:3334"),
		RoomAuthorizationTimeout:  authorizationTimeout,
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
