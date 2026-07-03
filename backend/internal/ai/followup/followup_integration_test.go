//go:build integration

package followup

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func TestSQLRepositoryLoadsFollowupContextAndCountsCap(t *testing.T) {
	db := newFollowupIntegrationDB(t)
	ctx := context.Background()
	postID := seedFollowupPost(t, db)
	parentID := seedFollowupComment(t, db, postID, nil, "AI", 1001, 0, "ai")
	replyID := seedFollowupComment(t, db, postID, &parentID, "USER", 0, 1, "user")
	for i := 0; i < 3; i++ {
		followupParentID := seedFollowupComment(t, db, postID, &parentID, "USER", 0, 1, fmt.Sprintf("user %d", i))
		if _, err := db.ExecContext(ctx, `
			INSERT INTO ai_reply_tasks (post_id, parent_comment_id, ai_agent_id, trigger_type, status)
			VALUES (?, ?, 1001, 'FOLLOWUP', 'SUCCESS')`, postID, followupParentID); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewSQLRepository(db)

	parent, err := repo.LoadComment(ctx, parentID)
	if err != nil {
		t.Fatal(err)
	}
	reply, err := repo.LoadComment(ctx, replyID)
	if err != nil {
		t.Fatal(err)
	}
	count, err := repo.CountFollowups(ctx, postID, 1001)
	if err != nil {
		t.Fatal(err)
	}

	if parent.CommentType != "AI" || parent.AIAgentID != 1001 || reply.CommentType != "USER" || reply.UserID != 1 || count != 3 {
		t.Fatalf("parent=%#v reply=%#v count=%d", parent, reply, count)
	}
}

func seedFollowupPost(t *testing.T, db *sqlx.DB) int64 {
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

func seedFollowupComment(t *testing.T, db *sqlx.DB, postID int64, parentID *int64, typ string, agentID, userID int64, content string) int64 {
	t.Helper()
	var userArg any
	if userID > 0 {
		userArg = userID
	}
	var agentArg any
	if agentID > 0 {
		agentArg = agentID
	}
	res, err := db.Exec(`
		INSERT INTO comments (post_id, user_id, parent_comment_id, comment_type, ai_agent_id, content)
		VALUES (?, ?, ?, ?, ?, ?)`, postID, userArg, parentID, typ, agentArg, content)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func newFollowupIntegrationDB(t *testing.T) *sqlx.DB {
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
