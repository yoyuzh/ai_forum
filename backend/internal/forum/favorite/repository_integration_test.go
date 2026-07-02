//go:build integration

package favorite

import (
	"context"
	"fmt"
	"os"
	"testing"

	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/outbox"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func TestSQLRepositoryFavoriteAndList(t *testing.T) {
	db := newFavoriteIntegrationDB(t)
	postID := seedPost(t, db)
	repo := NewSQLRepository()
	ctx := context.Background()

	changed, err := repo.Favorite(ctx, db, 1, postID)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("first favorite changed=false, want true")
	}
	changed, err = repo.Favorite(ctx, db, 1, postID)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("duplicate favorite changed=true, want false")
	}
	posts, err := repo.List(ctx, db, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0] != postID {
		t.Fatalf("favorites = %#v", posts)
	}
	changed, err = repo.Unfavorite(ctx, db, 1, postID)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("unfavorite changed=false, want true")
	}
	posts, err = repo.List(ctx, db, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatalf("favorites after unfavorite = %#v", posts)
	}
}

func TestServiceFavoriteOutboxRollsBackWithTransaction(t *testing.T) {
	db := newFavoriteIntegrationDB(t)
	postID := seedPost(t, db)
	svc := NewService(NewSQLRepository(), outbox.Append)
	ctx := context.Background()

	err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if err := svc.Favorite(ctx, tx, 1, postID); err != nil {
			return err
		}
		return fmt.Errorf("force rollback")
	})
	if err == nil {
		t.Fatal("expected rollback error")
	}

	var favorites, events int
	if err := db.GetContext(ctx, &favorites, `SELECT COUNT(*) FROM favorites WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &events, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ? AND event_type = ?`, postID, "post.favorited"); err != nil {
		t.Fatal(err)
	}
	if favorites != 0 || events != 0 {
		t.Fatalf("rollback left favorites=%d events=%d, want both 0", favorites, events)
	}
}

func seedPost(t *testing.T, db *sqlx.DB) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO posts (author_id, title, content, status) VALUES (1, 'p', 'body', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func newFavoriteIntegrationDB(t *testing.T) *sqlx.DB {
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
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name))
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
