package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/hibiken/asynq"
)

func TestPostCreatedConsumerEnqueuesTagPostOnce(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	consumer := NewPostCreatedConsumer(enqueuer)
	body, err := json.Marshal(EventEnvelope{EventID: "evt-1", EventType: "post.created", Payload: map[string]any{"post_id": float64(42)}})
	if err != nil {
		t.Fatal(err)
	}

	if err := consumer.Handle(context.Background(), body); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskType != TagPost || enqueuer.postID != 42 {
		t.Fatalf("enqueued = %s/%d, want tag_post/42", enqueuer.taskType, enqueuer.postID)
	}
}

func TestPostCreatedConsumerAcceptsEventEnvelopeWireShape(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	consumer := NewPostCreatedConsumer(enqueuer)
	body := []byte(`{"eventId":"evt-1","eventType":"post.created","aggregateType":"post","aggregateId":42,"occurredAt":"2026-07-02T00:00:00Z","payload":{"post_id":42}}`)

	if err := consumer.Handle(context.Background(), body); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskType != TagPost || enqueuer.postID != 42 {
		t.Fatalf("enqueued = %s/%d, want tag_post/42", enqueuer.taskType, enqueuer.postID)
	}
}

func TestPostCreatedConsumerSkipsAlreadyProcessedEvent(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	processed := &recordingProcessedStore{processed: true}
	consumer := NewPostCreatedConsumer(enqueuer, WithProcessedStore(processed, "worker.tag_post"))
	body, err := json.Marshal(EventEnvelope{EventID: "evt-1", EventType: "post.created", Payload: map[string]any{"post_id": float64(42)}})
	if err != nil {
		t.Fatal(err)
	}

	if err := consumer.Handle(context.Background(), body); err != nil {
		t.Fatal(err)
	}

	if enqueuer.calls != 0 {
		t.Fatalf("enqueue calls = %d, want 0", enqueuer.calls)
	}
}

func TestPostCreatedConsumerDoesNotMarkProcessedWhenEnqueueFails(t *testing.T) {
	enqueuer := &recordingEnqueuer{err: errors.New("redis down")}
	processed := &recordingProcessedStore{}
	consumer := NewPostCreatedConsumer(enqueuer, WithProcessedStore(processed, "worker.tag_post"))
	body, err := json.Marshal(EventEnvelope{EventID: "evt-1", EventType: "post.created", Payload: map[string]any{"post_id": float64(42)}})
	if err != nil {
		t.Fatal(err)
	}

	err = consumer.Handle(context.Background(), body)
	if err == nil {
		t.Fatal("expected enqueue error")
	}
	if processed.marked {
		t.Fatal("event must not be marked processed when enqueue fails")
	}
}

func TestPostTaggedConsumerEnqueuesDecideAIReply(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	consumer := NewPostTaggedConsumer(enqueuer)
	body, err := json.Marshal(EventEnvelope{EventID: "evt-2", EventType: "post.tagged", Payload: map[string]any{"post_id": float64(77)}})
	if err != nil {
		t.Fatal(err)
	}

	if err := consumer.Handle(context.Background(), body); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskType != DecideAIReply || enqueuer.postID != 77 {
		t.Fatalf("enqueued = %s/%d, want decide_ai_reply/77", enqueuer.taskType, enqueuer.postID)
	}
}

func TestSearchIndexConsumerEnqueuesSyncSearchIndex(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	consumer := NewSearchIndexConsumer(enqueuer)
	body := []byte(`{"eventId":"evt-search-1","eventType":"post.created","aggregateType":"post","aggregateId":42,"occurredAt":"2026-07-02T00:00:00Z","payload":{"post_id":42,"title":"ignored"}}`)

	if err := consumer.Handle(context.Background(), body); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskType != SyncSearchIndex || enqueuer.search.EventID != "evt-search-1" || enqueuer.search.EventType != "post.created" || enqueuer.search.PostID != 42 || enqueuer.search.Title != "ignored" {
		t.Fatalf("search task = %#v/%#v, want sync_search_index payload", enqueuer.taskType, enqueuer.search)
	}
}

func TestNotificationConsumerEnqueuesSendNotification(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	consumer := NewNotificationConsumer(enqueuer)
	body := []byte(`{"eventId":"evt-notify-1","eventType":"ai.reply.completed","aggregateType":"post","aggregateId":42,"occurredAt":"2026-07-02T00:00:00Z","payload":{"post_id":42,"comment_id":99}}`)

	if err := consumer.Handle(context.Background(), body); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskType != SendNotification || enqueuer.notify.EventID != "evt-notify-1" || enqueuer.notify.EventType != "ai.reply.completed" || enqueuer.notify.PostID != 42 || enqueuer.notify.CommentID != 99 {
		t.Fatalf("notification task = %#v/%#v, want send_notification payload", enqueuer.taskType, enqueuer.notify)
	}
}

func TestSQLProcessedStoreDelegatesToProcessedEventsHelpers(t *testing.T) {
	db := &processedDBTX{}
	store := NewSQLProcessedStore(db)

	processed, err := store.IsProcessed(context.Background(), "evt-1", "worker.tag_post")
	if err != nil {
		t.Fatal(err)
	}
	if !processed {
		t.Fatal("processed = false, want true")
	}
	if err := store.MarkProcessed(context.Background(), "evt-1", "worker.tag_post"); err != nil {
		t.Fatal(err)
	}
	if !db.marked {
		t.Fatal("expected MarkProcessed to write processed_events")
	}
}

type recordingEnqueuer struct {
	taskType        string
	postID          int64
	parentID        *int64
	parentCommentID int64
	replyCommentID  int64
	agentID         int64
	triggerType     string
	taskID          string
	maxRetry        int
	search          SyncSearchIndexPayload
	notify          SendNotificationPayload
	calls           int
	err             error
}

func (e *recordingEnqueuer) Enqueue(ctx context.Context, taskType string, payload any) error {
	e.calls++
	if e.err != nil {
		return e.err
	}
	e.taskType = taskType
	switch p := payload.(type) {
	case TagPostPayload:
		e.postID = p.PostID
	case DecideAIReplyPayload:
		e.postID = p.PostID
	case GenerateAIReplyPayload:
		e.postID = p.PostID
		e.parentID = p.ParentCommentID
		e.agentID = p.AIAgentID
		e.triggerType = p.TriggerType
	case JudgeAIFollowupPayload:
		e.postID = p.PostID
		e.parentCommentID = p.ParentCommentID
		e.replyCommentID = p.ReplyCommentID
	case SyncSearchIndexPayload:
		e.search = p
	case SendNotificationPayload:
		e.notify = p
	}
	return nil
}

func (e *recordingEnqueuer) EnqueueWithOptions(ctx context.Context, taskType string, payload any, opts ...asynq.Option) error {
	for _, opt := range opts {
		if opt.Type() == asynq.TaskIDOpt {
			e.taskID = opt.Value().(string)
		}
		if opt.Type() == asynq.MaxRetryOpt {
			e.maxRetry = opt.Value().(int)
		}
	}
	return e.Enqueue(ctx, taskType, payload)
}

type recordingProcessedStore struct {
	processed bool
	marked    bool
}

func (s *recordingProcessedStore) IsProcessed(context.Context, string, string) (bool, error) {
	return s.processed, nil
}

func (s *recordingProcessedStore) MarkProcessed(context.Context, string, string) error {
	s.marked = true
	return nil
}

type processedDBTX struct {
	marked bool
}

func (d *processedDBTX) ExecContext(_ context.Context, query string, _ ...interface{}) (sql.Result, error) {
	if strings.Contains(query, "INSERT INTO processed_events") {
		d.marked = true
	}
	return fakeResult(1), nil
}

func (d *processedDBTX) GetContext(_ context.Context, dest interface{}, _ string, _ ...interface{}) error {
	*(dest.(*int)) = 1
	return nil
}

func (d *processedDBTX) SelectContext(context.Context, interface{}, string, ...interface{}) error {
	return errors.New("unexpected select")
}

type fakeResult int64

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return int64(r), nil }
