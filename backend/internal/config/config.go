package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	JwtSecret     string
	DSN           string
	PubsubBrokers []string
	Environment   string
	CORSOrigins   []string
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := Config{
		Port:          "3333",
		DSN:           getEnv("DB_DSN", ""),
		PubsubBrokers: strings.Split(getEnv("PUBSUB_BROKERS", ""), ","),
		JwtSecret:     getEnv("JWT_SECRET", "SECRET"),
		Environment:   getEnv("ENV", "development"),
		CORSOrigins:   strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
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
