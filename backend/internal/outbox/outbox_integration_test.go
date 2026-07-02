//go:build integration

package outbox

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/database"
)

func TestAppendRollsBackWithCallerTransaction(t *testing.T) {
	db := newIntegrationDB(t)
	ctx := context.Background()

	err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if err := Append(ctx, tx, Event{
			EventID:       "evt-outbox-rollback",
			EventType:     "post.created",
			AggregateType: "post",
			AggregateID:   99,
			Payload:       map[string]any{"post_id": 99},
		}); err != nil {
			return err
		}
		return errors.New("rollback")
	})
	if err == nil {
		t.Fatal("expected rollback error")
	}

	var count int
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM outbox_events WHERE event_id = ?`, "evt-outbox-rollback"); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("outbox row count = %d, want 0 after rollback", count)
	}
}

func TestPublisherPublishesPendingRowAndMarksPublished(t *testing.T) {
	db := newIntegrationDB(t)
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO outbox_events (event_id, event_type, aggregate_type, aggregate_id, payload, status, created_at)
		VALUES (?, ?, ?, ?, ?, 'PENDING', NOW())`,
		"evt-publish-1", "post.created", "post", 42, `{"post_id":42}`)
	if err != nil {
		t.Fatal(err)
	}
	pub := &integrationPublisher{}
	publisher := NewPublisher(db, pub, Options{BatchSize: 100, MaxRetries: 3, ScanInterval: time.Hour})

	if err := publisher.ProcessOnce(ctx); err != nil {
		t.Fatal(err)
	}

	var status string
	if err := db.GetContext(ctx, &status, `SELECT status FROM outbox_events WHERE event_id = ?`, "evt-publish-1"); err != nil {
		t.Fatal(err)
	}
	if status != "PUBLISHED" || pub.routingKey != "post.created" {
		t.Fatalf("status/routing = %s/%s, want PUBLISHED/post.created", status, pub.routingKey)
	}
}

func newIntegrationDB(t *testing.T) *sqlx.DB {
	t.Helper()
	host := env("MYSQL_HOST", "127.0.0.1")
	port := env("MYSQL_PORT", "3306")
	user := env("MYSQL_USERNAME", "root")
	pass := env("MYSQL_PASSWORD", "ai_forum_root")
	name := env("MYSQL_DATABASE", "ai_forum")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, pass, host, port, name)

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

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type integrationPublisher struct {
	routingKey string
}

func (p *integrationPublisher) Publish(_ context.Context, _ string, routingKey string, _ []byte) error {
	p.routingKey = routingKey
	return nil
}
