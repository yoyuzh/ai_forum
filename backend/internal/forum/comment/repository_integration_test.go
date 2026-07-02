//go:build integration

package comment

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/outbox"
)

func TestSQLRepositoryCreateAndSoftDelete(t *testing.T) {
	db := newCommentIntegrationDB(t)
	ctx := context.Background()
	postID := seedPost(t, db)
	repo := NewSQLRepository()

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	c, err := repo.Create(ctx, tx, Comment{PostID: postID, UserID: 1, CommentType: "USER", Content: "hello"})
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := repo.IncrementCommentCount(ctx, tx, postID); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := repo.SoftDelete(ctx, tx, c.ID); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := repo.DecrementCommentCount(ctx, tx, postID); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.GetContext(ctx, &count, `SELECT comment_count FROM posts WHERE id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("comment_count = %d, want 0", count)
	}
}

func TestServiceCreateAndDeleteRollBackCommentAndOutbox(t *testing.T) {
	db := newCommentIntegrationDB(t)
	ctx := context.Background()
	postID := seedPost(t, db)
	svc := NewService(NewSQLRepository(), outbox.Append)

	err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if _, err := svc.Create(ctx, tx, CreateInput{PostID: postID, UserID: 1, Content: "rollback"}); err != nil {
			return err
		}
		return errors.New("rollback")
	})
	if err == nil {
		t.Fatal("expected create rollback error")
	}
	var comments int
	if err := db.GetContext(ctx, &comments, `SELECT COUNT(*) FROM comments WHERE content = ?`, "rollback"); err != nil {
		t.Fatal(err)
	}
	var createEvents int
	if err := db.GetContext(ctx, &createEvents, `SELECT COUNT(*) FROM outbox_events WHERE event_type = ?`, "comment.created"); err != nil {
		t.Fatal(err)
	}
	if comments != 0 || createEvents != 0 {
		t.Fatalf("comments/createEvents = %d/%d, want 0/0", comments, createEvents)
	}

	var commentID int64
	if err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		c, err := svc.Create(ctx, tx, CreateInput{PostID: postID, UserID: 1, Content: "keep"})
		if err != nil {
			return err
		}
		commentID = c.ID
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	err = database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if err := svc.Delete(ctx, tx, postID, commentID); err != nil {
			return err
		}
		return errors.New("rollback")
	})
	if err == nil {
		t.Fatal("expected delete rollback error")
	}
	var deleted int
	if err := db.GetContext(ctx, &deleted, `SELECT COUNT(*) FROM comments WHERE id = ? AND deleted_at IS NOT NULL`, commentID); err != nil {
		t.Fatal(err)
	}
	var deleteEvents int
	if err := db.GetContext(ctx, &deleteEvents, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ? AND event_type = ?`, commentID, "comment.deleted"); err != nil {
		t.Fatal(err)
	}
	if deleted != 0 || deleteEvents != 0 {
		t.Fatalf("deleted/deleteEvents = %d/%d, want 0/0", deleted, deleteEvents)
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

func newCommentIntegrationDB(t *testing.T) *sqlx.DB {
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
