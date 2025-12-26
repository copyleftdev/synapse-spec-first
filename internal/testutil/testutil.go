package testutil

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/synapse/synapse/internal/config"
	"github.com/synapse/synapse/internal/infra"
)

// TestInfra creates infrastructure connected to test containers
func TestInfra(ctx context.Context, t *testing.T, tc *TestContainers) (*infra.Infra, *config.Config) {
	t.Helper()

	natsURL, err := tc.NATSConnectionString(ctx)
	if err != nil {
		t.Fatalf("getting NATS connection string: %v", err)
	}

	postgresURL, err := tc.PostgresConnectionString(ctx)
	if err != nil {
		t.Fatalf("getting Postgres connection string: %v", err)
	}

	redisAddr, err := tc.RedisConnectionString(ctx)
	if err != nil {
		t.Fatalf("getting Redis connection string: %v", err)
	}

	cfg := &config.Config{
		HTTPPort:            8080,
		NATSURL:             natsURL,
		RedisAddr:           redisAddr,
		RedisPassword:       "",
		RedisDB:             0,
		PipelineConcurrency: 10,
		RetryMaxAttempts:    3,
		RetryBackoffMs:      100,
	}

	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		t.Fatalf("connecting to NATS: %v", err)
	}
	t.Cleanup(func() { nc.Close() })

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		t.Fatalf("connecting to Postgres: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("pinging Postgres: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("pinging Redis: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })

	return &infra.Infra{
		NATS:  nc,
		DB:    db,
		Redis: rdb,
	}, cfg
}
