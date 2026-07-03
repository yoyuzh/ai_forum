//go:build integration

package modelclient

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisTokenBucketLimiterAllowsBurstThenDenies(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     env("REDIS_ADDR", "127.0.0.1:6379"),
		Password: env("REDIS_PASSWORD", ""),
		DB:       9,
	})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	key := fmt.Sprintf("p7:limiter:%s", t.Name())
	t.Cleanup(func() { _ = client.Del(ctx, key+":tokens", key+":ts").Err() })
	now := time.Unix(100, 0)
	limiter := NewRedisTokenBucketLimiter(client, key, 1, 1, func() time.Time { return now })

	ok, err := limiter.Allow(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("first request denied")
	}
	ok, err = limiter.Allow(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("second request allowed, want rate limited")
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
