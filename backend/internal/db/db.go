package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lavianrose/flowforge/internal/config"
	"github.com/redis/go-redis/v9"
)

var (
	Pool *pgxpool.Pool
	RDB  *redis.Client
)

func Init(cfg *config.Config) error {
	// PostgreSQL
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.PostgresURL)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Pool = pool

	// Redis
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Password: cfg.RedisPwd,
		DB:       cfg.RedisDB,
	})

	for i := 0; i < 10; i++ {
		err = RDB.Ping(ctx).Err()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("unable to connect to redis: %w", err)
	}

	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
	if RDB != nil {
		RDB.Close()
	}
}
