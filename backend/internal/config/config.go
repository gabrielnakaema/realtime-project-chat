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
	godotenv.Load()

	env := getEnv("ENV", "development")

	port := getEnv("API_PORT", "3333")

	config := Config{
		Port:          port,
		DSN:           getEnv("DB_DSN", ""),
		PubsubBrokers: strings.Split(getEnv("PUBSUB_BROKERS", ""), ","),
		JwtSecret:     getEnv("JWT_SECRET", "SECRET"),
		Environment:   env,
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
