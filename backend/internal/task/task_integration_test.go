//go:build integration

// Integration tests for the task package, run against a live Redis container
// (docker-compose up -d redis). Build tag `integration` keeps these out of
// the default `go test ./...` run.
//
// Run with:
//
//	REDIS_ADDR=127.0.0.1:6379 \
//	go test -tags=integration ./internal/task/...
package task

import (
	"context"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/config"
)

// taskTestType is a local task type string used only by the P2 smoke test.
// Exported task type constants are deferred to P5; this string lives inside
// the test binary only.
const taskTestType = "p2:test"

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
	db, err := strconv.Atoi(get("REDIS_DB", "15"))
	if err != nil {
		db = 15
	}
	return config.RedisConfig{
		Addr:     get("REDIS_ADDR", "127.0.0.1:6379"),
		Password: get("REDIS_PASSWORD", ""),
		DB:       db,
	}
}

// TestAsynqRoundTrip verifies NewAsynqClient and NewAsynqServer both bind to
// the shared Redis broker and a trivial enqueue→process round-trip succeeds
// (spec: asynq-task-client, "Enqueue and process round-trip").
func TestAsynqRoundTrip(t *testing.T) {
	// Arrange — construct client + server against the same Redis broker.
	cfg := redisCfgFromEnv()
	client := NewAsynqClient(cfg)
	t.Cleanup(func() { _ = client.Close() })

	server := NewAsynqServer(cfg)
	t.Cleanup(server.Shutdown)

	// A buffered channel signals handler execution; wg guards against the
	// handler running more than once.
	var (
		gotPayload []byte
		mu         sync.Mutex
		done       = make(chan struct{}, 1)
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(taskTestType, func(_ context.Context, t *asynq.Task) error {
		mu.Lock()
		gotPayload = append([]byte(nil), t.Payload()...)
		mu.Unlock()
		select {
		case done <- struct{}{}:
		default:
		}
		return nil
	})

	// Act — start the server (non-blocking), then enqueue a test task.
	require.NoError(t, server.Start(mux), "server.Start must connect to Redis and begin processing")

	payload := []byte(`{"p":"p2"}`)
	info, err := client.Enqueue(asynq.NewTask(taskTestType, payload))
	require.NoError(t, err, "Enqueue must accept the test task")
	require.NotNil(t, info, "Enqueue must return non-nil TaskInfo")

	// Assert — the handler runs within the timeout and observes the payload.
	select {
	case <-done:
		// handler executed
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not execute within 5s timeout")
	}

	mu.Lock()
	assert.Equal(t, payload, gotPayload, "handler must receive the enqueued payload")
	mu.Unlock()
}
