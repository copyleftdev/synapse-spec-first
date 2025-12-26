package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/nats"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainers holds references to all test containers
type TestContainers struct {
	NATS     *nats.NATSContainer
	Postgres *postgres.PostgresContainer
	Redis    *redis.RedisContainer
}

// ContainerConfig holds configuration for test containers
type ContainerConfig struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
}

// DefaultConfig returns sensible defaults for testing
func DefaultConfig() *ContainerConfig {
	return &ContainerConfig{
		PostgresUser:     "synapse",
		PostgresPassword: "synapse",
		PostgresDB:       "synapse_test",
	}
}

// StartContainers starts all required test containers
func StartContainers(ctx context.Context, t *testing.T, cfg *ContainerConfig) (*TestContainers, error) {
	t.Helper()

	if cfg == nil {
		cfg = DefaultConfig()
	}

	tc := &TestContainers{}

	// Start NATS
	natsContainer, err := nats.Run(ctx, "nats:2.10-alpine")
	if err != nil {
		return nil, fmt.Errorf("starting NATS container: %w", err)
	}
	tc.NATS = natsContainer
	t.Cleanup(func() {
		if err := natsContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate NATS container: %v", err)
		}
	})

	// Start PostgreSQL
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(cfg.PostgresDB),
		postgres.WithUsername(cfg.PostgresUser),
		postgres.WithPassword(cfg.PostgresPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("starting Postgres container: %w", err)
	}
	tc.Postgres = postgresContainer
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate Postgres container: %v", err)
		}
	})

	// Start Redis
	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return nil, fmt.Errorf("starting Redis container: %w", err)
	}
	tc.Redis = redisContainer
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate Redis container: %v", err)
		}
	})

	return tc, nil
}

// NATSConnectionString returns the NATS connection string
func (tc *TestContainers) NATSConnectionString(ctx context.Context) (string, error) {
	return tc.NATS.ConnectionString(ctx)
}

// PostgresConnectionString returns the PostgreSQL connection string
func (tc *TestContainers) PostgresConnectionString(ctx context.Context) (string, error) {
	return tc.Postgres.ConnectionString(ctx, "sslmode=disable")
}

// RedisConnectionString returns the Redis connection string
func (tc *TestContainers) RedisConnectionString(ctx context.Context) (string, error) {
	host, err := tc.Redis.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := tc.Redis.MappedPort(ctx, "6379")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}
