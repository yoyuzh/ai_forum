//go:build integration

package decision

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
)

func TestSQLHandlerWritesDecisionLogsAndEnqueuesReply(t *testing.T) {
	db := newDecisionIntegrationDB(t)
	ctx := context.Background()
	res, err := db.ExecContext(ctx, `INSERT INTO posts (author_id, title, content, status) VALUES (1, 'AI debate', 'body', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	postID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO post_tags (post_id, tag_type, tag_name)
		VALUES (?, 'topic', 'debate'), (?, 'debate', 'high')`, postID, postID); err != nil {
		t.Fatal(err)
	}
	enqueuer := &recordingSQLReplyEnqueuer{}
	handler := NewSQLHandler(db, enqueuer)

	if err := handler.HandleDecideAIReply(ctx, postID); err != nil {
		t.Fatal(err)
	}

	var logs int
	if err := db.GetContext(ctx, &logs, `SELECT COUNT(*) FROM decision_logs WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if logs < 3 {
		t.Fatalf("decision logs = %d, want at least 3", logs)
	}
	if len(enqueuer.agentIDs) == 0 {
		t.Fatal("expected at least one generate_ai_reply enqueue")
	}
}

func TestSQLHandlerRedeliveryDoesNotDuplicateLogsOrTasks(t *testing.T) {
	db := newDecisionIntegrationDB(t)
	ctx := context.Background()
	res, err := db.ExecContext(ctx, `INSERT INTO posts (author_id, title, content, status) VALUES (1, 'AI debate', 'body', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	postID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO post_tags (post_id, tag_type, tag_name)
		VALUES (?, 'topic', 'debate'), (?, 'debate', 'high')`, postID, postID); err != nil {
		t.Fatal(err)
	}
	enqueuer := &recordingSQLReplyEnqueuer{}
	handler := NewSQLHandler(db, enqueuer)

	if err := handler.HandleDecideAIReply(ctx, postID); err != nil {
		t.Fatal(err)
	}
	firstEnqueued := len(enqueuer.agentIDs)
	if firstEnqueued == 0 {
		t.Fatal("expected first run to enqueue at least one generate_ai_reply task")
	}
	if err := handler.HandleDecideAIReply(ctx, postID); err != nil {
		t.Fatal(err)
	}

	var logs int
	if err := db.GetContext(ctx, &logs, `SELECT COUNT(*) FROM decision_logs WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if logs != 3 {
		t.Fatalf("decision logs = %d, want exactly 3", logs)
	}
	if len(enqueuer.agentIDs) != firstEnqueued {
		t.Fatalf("enqueued agents after redelivery = %d, want %d", len(enqueuer.agentIDs), firstEnqueued)
	}
}

func TestSQLHandlerFallbackEnqueuesWhenScoresAreLow(t *testing.T) {
	db := newDecisionIntegrationDB(t)
	ctx := context.Background()
	res, err := db.ExecContext(ctx, `INSERT INTO posts (author_id, title, content, status) VALUES (1, 'quiet post', 'body', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	postID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	enqueuer := &recordingSQLReplyEnqueuer{}
	handler := NewSQLHandler(db, enqueuer)

	if err := handler.HandleDecideAIReply(ctx, postID); err != nil {
		t.Fatal(err)
	}

	if len(enqueuer.agentIDs) != 1 || enqueuer.agentIDs[0] != 1001 {
		t.Fatalf("fallback enqueued agents = %#v, want [1001]", enqueuer.agentIDs)
	}
	var decision string
	if err := db.GetContext(ctx, &decision, `SELECT decision FROM decision_logs WHERE post_id = ? AND ai_agent_id = 1001`, postID); err != nil {
		t.Fatal(err)
	}
	if decision != DecisionFallback {
		t.Fatalf("fallback decision = %q, want %q", decision, DecisionFallback)
	}
}

func TestSQLHandlerRetriesEnqueueAfterLogsWereWritten(t *testing.T) {
	db := newDecisionIntegrationDB(t)
	ctx := context.Background()
	res, err := db.ExecContext(ctx, `INSERT INTO posts (author_id, title, content, status) VALUES (1, 'AI debate', 'body', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	postID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO post_tags (post_id, tag_type, tag_name)
		VALUES (?, 'topic', 'debate'), (?, 'debate', 'high')`, postID, postID); err != nil {
		t.Fatal(err)
	}
	enqueuer := &recordingSQLReplyEnqueuer{err: errors.New("redis down")}
	handler := NewSQLHandler(db, enqueuer)

	if err := handler.HandleDecideAIReply(ctx, postID); err == nil {
		t.Fatal("expected enqueue error")
	}
	enqueuer.err = nil
	if err := handler.HandleDecideAIReply(ctx, postID); err != nil {
		t.Fatal(err)
	}

	if len(enqueuer.agentIDs) == 0 {
		t.Fatal("expected retry to enqueue generate_ai_reply after prior enqueue failure")
	}
}

type recordingSQLReplyEnqueuer struct {
	agentIDs []int64
	seen     map[int64]bool
	err      error
}

func (e *recordingSQLReplyEnqueuer) EnqueueGenerateAIReply(_ context.Context, _ int64, agentID int64) error {
	if e.err != nil {
		return e.err
	}
	if e.seen == nil {
		e.seen = map[int64]bool{}
	}
	if e.seen[agentID] {
		return nil
	}
	e.seen[agentID] = true
	e.agentIDs = append(e.agentIDs, agentID)
	return nil
}

func newDecisionIntegrationDB(t *testing.T) *sqlx.DB {
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
