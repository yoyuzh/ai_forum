//go:build integration

package reply

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/ai/decision"
	"ai-forum/backend/internal/ai/modelclient"
	"ai-forum/backend/internal/ai/tagging"
	"ai-forum/backend/internal/database"
	forumpost "ai-forum/backend/internal/forum/post"
	"ai-forum/backend/internal/moderation"
	"ai-forum/backend/internal/outbox"
	"ai-forum/backend/internal/sse"
)

func TestSQLHandlerPersistsAICommentHotCounterAndCompletedOutbox(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postID := seedReplyPost(t, db)
	hot := &recordingHotTracker{}
	handler := NewSQLHandler(db, fakeModel{out: "AI says hello"}, allowModerator{}, allowLimiter{})
	handler.SetHotTracker(hot)

	if err := handler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"}); err != nil {
		t.Fatal(err)
	}

	var got struct {
		CommentType string `db:"comment_type"`
		AIAgentID   int64  `db:"ai_agent_id"`
		TriggerType string `db:"trigger_type"`
		Content     string `db:"content"`
	}
	if err := db.GetContext(ctx, &got, `SELECT comment_type, ai_agent_id, trigger_type, content FROM comments WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if got.CommentType != "AI" || got.AIAgentID != 1001 || got.TriggerType != "AUTO" || got.Content != "AI says hello" {
		t.Fatalf("comment = %#v", got)
	}

	var counts struct {
		CommentCount int `db:"comment_count"`
		AIReplyCount int `db:"ai_reply_count"`
	}
	if err := db.GetContext(ctx, &counts, `SELECT comment_count, ai_reply_count FROM posts WHERE id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if counts.CommentCount != 0 || counts.AIReplyCount != 0 {
		t.Fatalf("counts = %#v, want MySQL counters unchanged on hot path", counts)
	}
	if hot.postID != postID || hot.counter != HotCounterAIReply || hot.delta != 1 {
		t.Fatalf("hot = post %d counter %q delta %d", hot.postID, hot.counter, hot.delta)
	}

	var events int
	if err := db.GetContext(ctx, &events, `SELECT COUNT(*) FROM outbox_events WHERE event_type = 'ai.reply.completed' AND aggregate_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Fatalf("completed events = %d, want 1", events)
	}
}

func TestSQLHandlerBlocksModeratedReplyWithoutRetry(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postID := seedReplyPost(t, db)
	handler := NewSQLHandler(db, fakeModel{out: "blocked"}, blockModerator{}, allowLimiter{})

	if err := handler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"}); err != nil {
		t.Fatal(err)
	}

	var comments int
	if err := db.GetContext(ctx, &comments, `SELECT COUNT(*) FROM comments WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if comments != 0 {
		t.Fatalf("comments = %d, want 0", comments)
	}
	var status string
	if err := db.GetContext(ctx, &status, `SELECT status FROM ai_reply_tasks WHERE post_id = ? AND ai_agent_id = ?`, postID, int64(1001)); err != nil {
		t.Fatal(err)
	}
	if status != "BLOCKED" {
		t.Fatalf("status = %q, want BLOCKED", status)
	}
}

func TestSQLHandlerSkipsDuplicateSuccessfulAutoReply(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postID := seedReplyPost(t, db)
	handler := NewSQLHandler(db, fakeModel{out: "first"}, allowModerator{}, allowLimiter{})

	if err := handler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"}); err != nil {
		t.Fatal(err)
	}
	if err := handler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"}); err != nil {
		t.Fatal(err)
	}

	var comments int
	if err := db.GetContext(ctx, &comments, `SELECT COUNT(*) FROM comments WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if comments != 1 {
		t.Fatalf("comments = %d, want 1", comments)
	}
}

func TestSQLHandlerReturnsRetryOnRateLimit(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postID := seedReplyPost(t, db)
	handler := NewSQLHandler(db, fakeModel{out: "unused"}, allowModerator{}, denyLimiter{})

	err := handler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"})
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}

func TestSQLHandlerWritesFailedOutboxOnModelFailure(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postID := seedReplyPost(t, db)
	handler := NewSQLHandler(db, fakeModel{err: errors.New("model down")}, allowModerator{}, allowLimiter{})

	if err := handler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"}); err == nil {
		t.Fatal("expected model error")
	}

	var events int
	if err := db.GetContext(ctx, &events, `SELECT COUNT(*) FROM outbox_events WHERE event_type = 'ai.reply.failed' AND aggregate_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Fatalf("failed events = %d, want 1", events)
	}
}

func TestSQLHandlerSkipsDisabledAgentWithoutModelCall(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postID := seedReplyPost(t, db)
	if _, err := db.ExecContext(ctx, `UPDATE ai_agents SET enabled = FALSE WHERE id = ?`, int64(1001)); err != nil {
		t.Fatal(err)
	}
	handler := NewSQLHandler(db, fakeModel{out: "should not be used"}, allowModerator{}, allowLimiter{})

	if err := handler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"}); err != nil {
		t.Fatal(err)
	}

	var comments int
	if err := db.GetContext(ctx, &comments, `SELECT COUNT(*) FROM comments WHERE post_id = ?`, postID); err != nil {
		t.Fatal(err)
	}
	if comments != 0 {
		t.Fatalf("comments = %d, want 0", comments)
	}
	var status string
	if err := db.GetContext(ctx, &status, `SELECT status FROM ai_reply_tasks WHERE post_id = ? AND ai_agent_id = ?`, postID, int64(1001)); err != nil {
		t.Fatal(err)
	}
	if status != "SKIPPED" {
		t.Fatalf("status = %q, want SKIPPED", status)
	}
}

func TestSQLHandlerNotifiesAfterSuccessfulReply(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postID := seedReplyPost(t, db)
	notifier := &recordingNotifier{}
	handler := NewSQLHandler(db, fakeModel{out: "AI says hello"}, allowModerator{}, allowLimiter{})
	handler.SetNotifier(notifier)

	if err := handler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"}); err != nil {
		t.Fatal(err)
	}

	if notifier.postID != postID || notifier.eventType != "ai_reply_completed" {
		t.Fatalf("notifier = post %d type %q", notifier.postID, notifier.eventType)
	}
}

func TestFullChainPostWriteDecisionReplySSEAndStatus(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postSvc := forumpost.NewService(forumpost.NewSQLRepository(), outbox.Append)
	var post forumpost.Post
	if err := database.RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		var err error
		post, err = postSvc.CreatePost(ctx, tx, forumpost.CreateInput{
			AuthorID: 1,
			Title:    "AI risk debate",
			Content:  "Should we discuss safety tradeoffs?",
		})
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var postCreated int
	if err := db.GetContext(ctx, &postCreated, `SELECT COUNT(*) FROM outbox_events WHERE event_type = 'post.created' AND aggregate_id = ?`, post.ID); err != nil {
		t.Fatal(err)
	}
	if postCreated != 1 {
		t.Fatalf("post.created events = %d, want 1", postCreated)
	}

	if err := tagging.NewSQLHandler(db, tagging.RuleTagger{}).HandleTagPost(ctx, post.ID); err != nil {
		t.Fatal(err)
	}
	var postTagged int
	if err := db.GetContext(ctx, &postTagged, `SELECT COUNT(*) FROM outbox_events WHERE event_type = 'post.tagged' AND aggregate_id = ?`, post.ID); err != nil {
		t.Fatal(err)
	}
	if postTagged != 1 {
		t.Fatalf("post.tagged events = %d, want 1", postTagged)
	}

	enqueuer := &recordingDecisionReplyEnqueuer{}
	if err := decision.NewSQLHandler(db, enqueuer).HandleDecideAIReply(ctx, post.ID); err != nil {
		t.Fatal(err)
	}
	if len(enqueuer.agentIDs) == 0 {
		t.Fatal("decision did not enqueue generate_ai_reply")
	}

	hub := sse.NewHub()
	events, cancel := hub.Subscribe(post.ID)
	defer cancel()
	replyHandler := NewSQLHandler(db, fakeModel{out: "AI chain reply"}, allowModerator{}, allowLimiter{})
	replyHandler.SetNotifier(hubNotifier{hub: hub})
	if err := replyHandler.HandleGenerateAIReply(ctx, Task{PostID: post.ID, AgentID: enqueuer.agentIDs[0], TriggerType: "AUTO"}); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Type != "ai_reply_completed" {
			t.Fatalf("event = %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ai_reply_completed")
	}

	status, err := sse.NewSQLStatusStore(db).AIStatus(ctx, post.ID)
	if err != nil {
		t.Fatal(err)
	}
	if status.CompletedCount != 1 || status.RunningCount != 0 {
		t.Fatalf("ai status = %#v, want completed=1 running=0", status)
	}
}

func TestConcurrentDuplicatePostTaggedAndGenerateAIReplyProduceOneDecisionAndComment(t *testing.T) {
	db := newReplyIntegrationDB(t)
	ctx := context.Background()
	postID := seedReplyPost(t, db)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO post_tags (post_id, tag_type, tag_name)
		VALUES (?, 'topic', 'debate'), (?, 'debate', 'high')`, postID, postID); err != nil {
		t.Fatal(err)
	}

	enqueuer := &recordingDecisionReplyEnqueuer{}
	decisionHandler := decision.NewSQLHandler(db, enqueuer)
	runConcurrent(t, 12, func() error {
		return decisionHandler.HandleDecideAIReply(ctx, postID)
	})

	replyHandler := NewSQLHandler(db, fakeModel{out: "one concurrent reply"}, allowModerator{}, allowLimiter{})
	runConcurrent(t, 12, func() error {
		return replyHandler.HandleGenerateAIReply(ctx, Task{PostID: postID, AgentID: 1001, TriggerType: "AUTO"})
	})

	var decisions int
	if err := db.GetContext(ctx, &decisions, `
		SELECT COUNT(*) FROM decision_logs
		WHERE post_id = ? AND ai_agent_id = 1001 AND trigger_type = 'AUTO'`, postID); err != nil {
		t.Fatal(err)
	}
	if decisions != 1 {
		t.Fatalf("decisions = %d, want 1", decisions)
	}

	var comments int
	if err := db.GetContext(ctx, &comments, `
		SELECT COUNT(*) FROM comments
		WHERE post_id = ? AND ai_agent_id = 1001 AND trigger_type = 'AUTO'`, postID); err != nil {
		t.Fatal(err)
	}
	if comments != 1 {
		t.Fatalf("comments = %d, want 1", comments)
	}
}

type fakeModel struct {
	out string
	err error
}

func (m fakeModel) Generate(context.Context, modelclient.Request) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.out, nil
}

type allowModerator struct{}

func (allowModerator) Review(context.Context, moderation.Input) (moderation.Result, error) {
	return moderation.Result{Allowed: true}, nil
}

type blockModerator struct{}

func (blockModerator) Review(context.Context, moderation.Input) (moderation.Result, error) {
	return moderation.Result{Allowed: false, Reason: "blocked"}, nil
}

type allowLimiter struct{}

func (allowLimiter) Allow(context.Context) (bool, error) { return true, nil }

type denyLimiter struct{}

func (denyLimiter) Allow(context.Context) (bool, error) { return false, nil }

type recordingNotifier struct {
	postID    int64
	eventType string
}

func (n *recordingNotifier) Notify(_ context.Context, postID int64, event Event) error {
	n.postID = postID
	n.eventType = event.Type
	return nil
}

type recordingHotTracker struct {
	postID  int64
	counter HotCounter
	delta   int64
}

func (h *recordingHotTracker) RecordInteraction(_ context.Context, postID int64, counter HotCounter, delta int64) error {
	h.postID = postID
	h.counter = counter
	h.delta = delta
	return nil
}

type recordingDecisionReplyEnqueuer struct {
	agentIDs []int64
}

func (e *recordingDecisionReplyEnqueuer) EnqueueAutoGenerateAIReply(_ context.Context, _ int64, agentID int64) error {
	e.agentIDs = append(e.agentIDs, agentID)
	return nil
}

type hubNotifier struct {
	hub *sse.InMemoryHub
}

func (n hubNotifier) Notify(ctx context.Context, postID int64, event Event) error {
	return n.hub.Publish(ctx, postID, event)
}

func runConcurrent(t *testing.T, n int, fn func() error) {
	t.Helper()
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- fn()
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func newReplyIntegrationDB(t *testing.T) *sqlx.DB {
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

func seedReplyPost(t *testing.T, db *sqlx.DB) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO posts (author_id, title, content, status) VALUES (1, 'reply post', 'body', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
