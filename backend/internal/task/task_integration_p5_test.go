//go:build integration

package task

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func TestCleanupProcessedEventsRowsDeletesOnlyOldRows(t *testing.T) {
	db := newTaskIntegrationDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx,
		`INSERT INTO processed_events (event_id, consumer_name, processed_at) VALUES (?, ?, ?), (?, ?, ?)`,
		"evt-old", "c", time.Now().Add(-31*24*time.Hour),
		"evt-new", "c", time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := CleanupProcessedEventsRows(ctx, db); err != nil {
		t.Fatal(err)
	}

	var oldCount, newCount int
	if err := db.GetContext(ctx, &oldCount, `SELECT COUNT(*) FROM processed_events WHERE event_id = 'evt-old'`); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &newCount, `SELECT COUNT(*) FROM processed_events WHERE event_id = 'evt-new'`); err != nil {
		t.Fatal(err)
	}
	if oldCount != 0 || newCount != 1 {
		t.Fatalf("old/new counts = %d/%d, want 0/1", oldCount, newCount)
	}
}

func newTaskIntegrationDB(t *testing.T) *sqlx.DB {
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

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
