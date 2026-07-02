//go:build integration

package post

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

func TestSQLRepositoryCreateAndUpdateStatus(t *testing.T) {
	db := newPostIntegrationDB(t)
	ctx := context.Background()
	repo := NewSQLRepository()

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	created, err := repo.Create(ctx, tx, Post{AuthorID: 1, Title: "hello", Content: "body", Status: "NORMAL"})
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := repo.UpdateStatus(ctx, tx, created.ID, "HIDDEN"); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	var status string
	if err := db.GetContext(ctx, &status, `SELECT status FROM posts WHERE id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if status != "HIDDEN" {
		t.Fatalf("status = %q, want HIDDEN", status)
	}
}

func TestSQLRepositoryReadUpdateAndSoftDelete(t *testing.T) {
	db := newPostIntegrationDB(t)
	ctx := context.Background()
	repo := NewSQLRepository()

	var postID int64
	if err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		created, err := repo.Create(ctx, tx, Post{AuthorID: 1, Title: "read", Content: "body", Status: "NORMAL"})
		if err != nil {
			return err
		}
		postID = created.ID
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		listed, err := repo.List(ctx, tx)
		if err != nil {
			return err
		}
		if len(listed) != 1 || listed[0].ID != postID {
			t.Fatalf("listed = %#v", listed)
		}
		got, err := repo.Get(ctx, tx, postID)
		if err != nil {
			return err
		}
		if got.Title != "read" {
			t.Fatalf("got = %#v", got)
		}
		updated, err := repo.Update(ctx, tx, Post{ID: postID, AuthorID: 1, Title: "updated", Content: "new body"})
		if err != nil {
			return err
		}
		if updated.Title != "updated" {
			t.Fatalf("updated = %#v", updated)
		}
		return repo.SoftDelete(ctx, tx, postID)
	}); err != nil {
		t.Fatal(err)
	}

	var visible int
	if err := db.GetContext(ctx, &visible, `SELECT COUNT(*) FROM posts WHERE id = ? AND deleted_at IS NULL`, postID); err != nil {
		t.Fatal(err)
	}
	if visible != 0 {
		t.Fatalf("visible = %d, want 0 after soft delete", visible)
	}
}

func TestServiceCreatePostRollsBackPostAndOutbox(t *testing.T) {
	db := newPostIntegrationDB(t)
	ctx := context.Background()
	svc := NewService(NewSQLRepository(), outbox.Append)

	err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if _, err := svc.CreatePost(ctx, tx, CreateInput{AuthorID: 1, Title: "rollback", Content: "body"}); err != nil {
			return err
		}
		return errors.New("rollback")
	})
	if err == nil {
		t.Fatal("expected rollback error")
	}
	var posts int
	if err := db.GetContext(ctx, &posts, `SELECT COUNT(*) FROM posts WHERE title = ?`, "rollback"); err != nil {
		t.Fatal(err)
	}
	var events int
	if err := db.GetContext(ctx, &events, `SELECT COUNT(*) FROM outbox_events WHERE event_type = ?`, "post.created"); err != nil {
		t.Fatal(err)
	}
	if posts != 0 || events != 0 {
		t.Fatalf("posts/events = %d/%d, want 0/0", posts, events)
	}
}

func TestServiceUpdateStatusRollsBackPostAndOutbox(t *testing.T) {
	db := newPostIntegrationDB(t)
	ctx := context.Background()
	repo := NewSQLRepository()
	svc := NewService(repo, outbox.Append)

	var postID int64
	if err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		created, err := repo.Create(ctx, tx, Post{AuthorID: 1, Title: "status rollback", Content: "body", Status: "NORMAL"})
		if err != nil {
			return err
		}
		postID = created.ID
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if err := svc.UpdateStatus(ctx, tx, postID, "HIDDEN"); err != nil {
			return err
		}
		return errors.New("rollback")
	})
	if err == nil {
		t.Fatal("expected rollback error")
	}
	var status string
	if err := db.GetContext(ctx, &status, `SELECT status FROM posts WHERE id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	var events int
	if err := db.GetContext(ctx, &events, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ? AND event_type = ?`, postID, "post.moderated"); err != nil {
		t.Fatal(err)
	}
	if status != "NORMAL" || events != 0 {
		t.Fatalf("status/events = %q/%d, want NORMAL/0", status, events)
	}
}

func TestServiceUpdateAndDeleteRollBackPostAndOutbox(t *testing.T) {
	db := newPostIntegrationDB(t)
	ctx := context.Background()
	repo := NewSQLRepository()
	svc := NewService(repo, outbox.Append)

	var postID int64
	if err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		created, err := repo.Create(ctx, tx, Post{AuthorID: 1, Title: "original", Content: "body", Status: "NORMAL"})
		if err != nil {
			return err
		}
		postID = created.ID
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if _, err := svc.UpdateOwn(ctx, tx, UpdateInput{PostID: postID, AuthorID: 1, Title: "rolled back", Content: "body"}); err != nil {
			return err
		}
		return errors.New("rollback")
	})
	if err == nil {
		t.Fatal("expected update rollback error")
	}
	var title string
	if err := db.GetContext(ctx, &title, `SELECT title FROM posts WHERE id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	var updateEvents int
	if err := db.GetContext(ctx, &updateEvents, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ? AND event_type = ?`, postID, "post.updated"); err != nil {
		t.Fatal(err)
	}
	if title != "original" || updateEvents != 0 {
		t.Fatalf("title/updateEvents = %q/%d, want original/0", title, updateEvents)
	}

	err = database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if err := svc.Delete(ctx, tx, postID); err != nil {
			return err
		}
		return errors.New("rollback")
	})
	if err == nil {
		t.Fatal("expected delete rollback error")
	}
	var deleted int
	if err := db.GetContext(ctx, &deleted, `SELECT COUNT(*) FROM posts WHERE id = ? AND deleted_at IS NOT NULL`, postID); err != nil {
		t.Fatal(err)
	}
	var deleteEvents int
	if err := db.GetContext(ctx, &deleteEvents, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ? AND event_type = ?`, postID, "post.deleted"); err != nil {
		t.Fatal(err)
	}
	if deleted != 0 || deleteEvents != 0 {
		t.Fatalf("deleted/deleteEvents = %d/%d, want 0/0", deleted, deleteEvents)
	}
}

func newPostIntegrationDB(t *testing.T) *sqlx.DB {
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
