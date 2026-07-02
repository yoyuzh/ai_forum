//go:build integration

// Integration tests for the cache package, run against a live Redis container
// (docker-compose up -d redis). Build tag `integration` keeps these out of the
// default `go test ./...` run.
//
// Run with:
//
//	REDIS_ADDR=127.0.0.1:6379 REDIS_PASSWORD= REDIS_DB=0 \
//	go test -tags=integration ./internal/cache/...
package cache

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/config"
)

// redisCfgFromEnv builds the Redis config the same way the loader does, from
// the same env vars. Defaults match docker-compose so `docker compose up -d`
// + `go test -tags=integration` works out of the box.
func redisCfgFromEnv() config.RedisConfig {
	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &db); err != nil {
			db = 0
		}
	}

	return config.RedisConfig{
		Addr:     get("REDIS_ADDR", "127.0.0.1:6379"),
		Password: get("REDIS_PASSWORD", ""),
		DB:       db,
	}
}

// newTestRedis returns a connected *redis.Client. The client is closed via
// t.Cleanup so each test gets an isolated connection.
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	cfg := redisCfgFromEnv()

	client, err := NewRedis(cfg)
	require.NoError(t, err, "NewRedis must connect to the live Redis container")
	t.Cleanup(func() { _ = client.Close() })

	return client
}

// TestRedisPing verifies NewRedis produces a client whose Ping succeeds
// (spec: redis-client, "Ping SHALL succeed against the configured Redis").
func TestRedisPing(t *testing.T) {
	client := newTestRedis(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, client.Ping(ctx).Err(), "Ping must succeed against live Redis")
}

// TestRedisRoundTrip verifies a Set/Get round-trip returns the stored value
// (spec: redis-client, "client.Set/Get of a test key round-trips successfully").
func TestRedisRoundTrip(t *testing.T) {
	client := newTestRedis(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Arrange: unique key scoped to this test run, short TTL so stale keys
	// do not survive a flaky re-run.
	key := fmt.Sprintf("p2:redis:test:%s", t.Name())
	value := "round-trip-value"
	ttl := 10 * time.Second

	// Act
	require.NoError(t, client.Set(ctx, key, value, ttl).Err(),
		"Set must write the test key")

	got, err := client.Get(ctx, key).Result()
	require.NoError(t, err, "Get must read the test key back")

	// Assert
	assert.Equal(t, value, got, "Get must return the value written by Set")

	// Cleanup: delete the key so the test leaves no residue.
	t.Cleanup(func() {
		delCtx, delCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer delCancel()
		_ = client.Del(delCtx, key)
	})
}
