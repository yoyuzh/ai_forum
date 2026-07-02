//go:build integration

package tagging

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func TestSQLHandlerWritesTagsAndPostTaggedOutbox(t *testing.T) {
	db := newTaggingIntegrationDB(t)
	ctx := context.Background()
	res, err := db.ExecContext(ctx, `INSERT INTO posts (author_id, title, content, status) VALUES (1, 'AI risk debate', 'Should we discuss safety?', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	postID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	handler := NewSQLHandler(db, RuleTagger{})
	if err := handler.HandleTagPost(ctx, postID); err != nil {
		t.Fatal(err)
	}

	var tagTypes int
	if err := db.GetContext(ctx, &tagTypes, `SELECT COUNT(DISTINCT tag_type) FROM post_tags WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if tagTypes != 5 {
		t.Fatalf("tag types = %d, want 5", tagTypes)
	}
	var outboxRows int
	if err := db.GetContext(ctx, &outboxRows, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ? AND event_type = 'post.tagged'`, postID); err != nil {
		t.Fatal(err)
	}
	if outboxRows != 1 {
		t.Fatalf("post.tagged outbox rows = %d, want 1", outboxRows)
	}
}

func newTaggingIntegrationDB(t *testing.T) *sqlx.DB {
	t.Helper()
	host := env("MYSQL_HOST", "127.0.0.1")
	port := env("MYSQL_PORT", "3306")
	user := env("MYSQL_USERNAME", "root")
	pass := env("MYSQL_PASSWORD", "ai_forum_root")
	name := env("MYSQL_DATABASE", "ai_forum")

	m, err := migrate.New("file://../../../migrations", "mysql://"+user+":"+pass+"@tcp("+host+":"+port+")/"+name)
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
