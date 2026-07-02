//go:build integration

package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/mq"
	"ai-forum/backend/internal/outbox"
	"ai-forum/backend/internal/task"
)

func TestWorkerConsumesPostCreatedThroughDecisionChain(t *testing.T) {
	db := newWorkerIntegrationDB(t)
	ctx := context.Background()

	rabbit, err := mq.NewRabbitMQ(config.RabbitMQConfig{URL: env("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rabbit.Close() })
	if err := mq.DeclareTopology(rabbit); err != nil {
		t.Fatal(err)
	}
	redisCfg := config.RedisConfig{Addr: env("REDIS_ADDR", "127.0.0.1:6379")}
	redisCfg.DB = 9
	client := task.NewAsynqClient(redisCfg)
	t.Cleanup(func() { _ = client.Close() })
	server := task.NewAsynqServer(redisCfg)
	t.Cleanup(server.Shutdown)

	app := &App{DB: db, RabbitMQ: rabbit, AsynqClient: client, AsynqServer: server}
	worker := app.NewWorker()
	if err := worker.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := worker.Stop(stopCtx); err != nil {
			t.Fatal(err)
		}
	})
	waitForRabbitConsumers(t, rabbit, map[string]int{mq.QueuePostTagging: 1, mq.QueueAIDecision: 1})

	res, err := db.ExecContext(ctx, `INSERT INTO posts (author_id, title, content, status) VALUES (1, 'AI risk debate', 'Should we discuss safety?', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	postID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	pub := outbox.NewPublisher(db, mq.NewPublisher(rabbit), outbox.Options{ScanInterval: time.Hour})
	if _, err := db.ExecContext(ctx, `
		INSERT INTO outbox_events (event_id, event_type, aggregate_type, aggregate_id, payload, status, created_at)
		VALUES (?, 'post.created', 'post', ?, ?, 'PENDING', NOW())`,
		"evt-p6-chain-1", postID, mustJSON(t, map[string]any{"post_id": postID})); err != nil {
		t.Fatal(err)
	}
	if err := pub.ProcessOnce(ctx); err != nil {
		t.Fatal(err)
	}

	waitFor(t, 10*time.Second, func() bool {
		var count int
		return db.GetContext(ctx, &count, `SELECT COUNT(*) FROM outbox_events WHERE event_type = 'post.tagged' AND aggregate_id = ?`, postID) == nil && count == 1
	})
	if err := pub.ProcessOnce(ctx); err != nil {
		t.Fatal(err)
	}
	waitFor(t, 10*time.Second, func() bool {
		var count int
		return db.GetContext(ctx, &count, `SELECT COUNT(*) FROM decision_logs WHERE post_id = ?`, postID) == nil && count >= 3
	})
}

func waitForRabbitConsumers(t *testing.T, rabbit *mq.Connection, want map[string]int) {
	t.Helper()
	waitFor(t, 5*time.Second, func() bool {
		ch, err := rabbit.Channel()
		if err != nil {
			return false
		}
		defer ch.Close()
		for queue, consumers := range want {
			q, err := ch.QueueInspect(queue)
			if err != nil || q.Consumers < consumers {
				return false
			}
		}
		return true
	})
}

func newWorkerIntegrationDB(t *testing.T) *sqlx.DB {
	t.Helper()
	host := env("MYSQL_HOST", "127.0.0.1")
	port := env("MYSQL_PORT", "3306")
	user := env("MYSQL_USERNAME", "root")
	pass := env("MYSQL_PASSWORD", "ai_forum_root")
	name := env("MYSQL_DATABASE", "ai_forum")

	m, err := migrate.New("file://../../migrations", "mysql://"+user+":"+pass+"@tcp("+host+":"+port+")/"+name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = m.Close() })
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}

	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, pass, host, port, name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func waitFor(t *testing.T, timeout time.Duration, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
