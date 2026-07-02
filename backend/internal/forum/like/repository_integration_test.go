//go:build integration

package like

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

func TestSQLRepositoryLikeDuplicateAndUnlike(t *testing.T) {
	db := newLikeIntegrationDB(t)
	postID := seedPost(t, db)
	repo := NewSQLRepository()
	ctx := context.Background()

	changed, err := repo.Like(ctx, db, 1, postID)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("first like changed=false, want true")
	}
	changed, err = repo.Like(ctx, db, 1, postID)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("duplicate like changed=true, want false")
	}
	count, err := repo.Count(ctx, db, postID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
	changed, err = repo.Unlike(ctx, db, 1, postID)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("unlike changed=false, want true")
	}
	count, err = repo.Count(ctx, db, postID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("count after unlike = %d, want 0", count)
	}
}

func TestServiceLikeOutboxRollsBackWithTransaction(t *testing.T) {
	db := newLikeIntegrationDB(t)
	postID := seedPost(t, db)
	svc := NewService(NewSQLRepository(), outbox.Append)
	ctx := context.Background()

	err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if err := svc.Like(ctx, tx, 1, postID); err != nil {
			return err
		}
		return fmt.Errorf("force rollback")
	})
	if err == nil {
		t.Fatal("expected rollback error")
	}

	var likes, events int
	if err := db.GetContext(ctx, &likes, `SELECT COUNT(*) FROM likes WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &events, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ? AND event_type = ?`, postID, "post.liked"); err != nil {
		t.Fatal(err)
	}
	if likes != 0 || events != 0 {
		t.Fatalf("rollback left likes=%d events=%d, want both 0", likes, events)
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

func newLikeIntegrationDB(t *testing.T) *sqlx.DB {
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
