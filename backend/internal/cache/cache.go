// Package cache provides the Redis client used for cache, rate limiting,
// hot-score counters, and the Asynq broker connection. P2 only proves the
// connection; key schemes and counter helpers land in later phases.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ai-forum/backend/internal/config"
)

// pingTimeout is the maximum time NewRedis will wait for a Ping round-trip
// before reporting the client as not ready.
const pingTimeout = 5 * time.Second

// NewRedis returns a Redis client that has been verified reachable via Ping.
// Construction dials and pings so the readiness path is meaningful: callers
// can assume a returned client is connected. Redis key conventions, rate
// limiting, and counter helpers are intentionally NOT defined here (P7/P10).
func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}
