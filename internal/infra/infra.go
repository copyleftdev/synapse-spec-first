package infra

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/synapse/synapse/internal/config"
)

// Infra holds all infrastructure connections
type Infra struct {
	NATS   *nats.Conn
	DB     *sql.DB
	Redis  *redis.Client
	Config *config.Config
}

// New creates a new Infra instance with all connections
func New(ctx context.Context, cfg *config.Config) (*Infra, error) {
	infra := &Infra{Config: cfg}

	// Connect to NATS
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}
	infra.NATS = nc

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", cfg.PostgresDSN())
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		nc.Close()
		db.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	infra.DB = db

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		nc.Close()
		db.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}
	infra.Redis = rdb

	return infra, nil
}

// Close closes all infrastructure connections
func (i *Infra) Close() {
	if i.NATS != nil {
		i.NATS.Close()
	}
	if i.DB != nil {
		i.DB.Close()
	}
	if i.Redis != nil {
		i.Redis.Close()
	}
}

// Healthy returns true if all connections are healthy
func (i *Infra) Healthy(ctx context.Context) map[string]error {
	results := make(map[string]error)

	// Check NATS
	if i.NATS != nil && i.NATS.IsConnected() {
		results["nats"] = nil
	} else {
		results["nats"] = fmt.Errorf("not connected")
	}

	// Check PostgreSQL
	if i.DB != nil {
		results["postgres"] = i.DB.PingContext(ctx)
	} else {
		results["postgres"] = fmt.Errorf("not connected")
	}

	// Check Redis
	if i.Redis != nil {
		results["redis"] = i.Redis.Ping(ctx).Err()
	} else {
		results["redis"] = fmt.Errorf("not connected")
	}

	return results
}
