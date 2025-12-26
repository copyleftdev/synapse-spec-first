package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// HTTP server
	HTTPPort int

	// NATS
	NATSURL string

	// PostgreSQL
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Pipeline
	PipelineConcurrency int
	RetryMaxAttempts    int
	RetryBackoffMs      int
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:            getEnvInt("HTTP_PORT", 8080),
		NATSURL:             getEnv("NATS_URL", "nats://localhost:4222"),
		PostgresHost:        getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:        getEnvInt("POSTGRES_PORT", 5432),
		PostgresUser:        getEnv("POSTGRES_USER", "synapse"),
		PostgresPassword:    getEnv("POSTGRES_PASSWORD", "synapse"),
		PostgresDB:          getEnv("POSTGRES_DB", "synapse"),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		RedisDB:             getEnvInt("REDIS_DB", 0),
		PipelineConcurrency: getEnvInt("PIPELINE_CONCURRENCY", 10),
		RetryMaxAttempts:    getEnvInt("RETRY_MAX_ATTEMPTS", 3),
		RetryBackoffMs:      getEnvInt("RETRY_BACKOFF_MS", 1000),
	}

	return cfg, nil
}

// PostgresDSN returns the PostgreSQL connection string
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
