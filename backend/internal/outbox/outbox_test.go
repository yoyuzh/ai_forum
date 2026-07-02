package outbox

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAppendInsertsPendingEventWithoutPublish(t *testing.T) {
	db := &recordingDBTX{}
	err := Append(context.Background(), db, Event{
		EventType:     "post.created",
		AggregateType: "post",
		AggregateID:   42,
		Payload:       map[string]any{"post_id": 42},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(strings.ToLower(db.query), "insert into outbox_events") {
		t.Fatalf("query must insert into outbox_events, got %q", db.query)
	}
	for _, want := range []string{"event_id", "event_type", "aggregate_type", "aggregate_id", "payload", "status", "created_at"} {
		if !strings.Contains(strings.ToLower(db.query), want) {
			t.Fatalf("query missing %s: %s", want, db.query)
		}
	}
	if got := db.args[5]; got != "PENDING" {
		t.Fatalf("status = %v, want PENDING", got)
	}
	if got := db.args[1]; got != "post.created" {
		t.Fatalf("event_type arg = %v", got)
	}
	if got := db.args[3]; got != int64(42) {
		t.Fatalf("aggregate_id arg = %v", got)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(db.args[4].(string)), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["post_id"].(float64) != 42 {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestPublisherProcessOnceMarksPublished(t *testing.T) {
	db := &recordingDBTX{
		pendingRows: []Record{{
			ID:            7,
			EventID:       "evt-7",
			EventType:     "post.created",
			AggregateType: "post",
			AggregateID:   42,
			Payload:       json.RawMessage(`{"post_id":42}`),
		}},
	}
	pub := &recordingPublisher{}
	publisher := NewPublisher(db, pub, Options{BatchSize: 100, MaxRetries: 3, ScanInterval: time.Hour})

	if err := publisher.ProcessOnce(context.Background()); err != nil {
		t.Fatal(err)
	}

	if pub.routingKey != "post.created" {
		t.Fatalf("routing key = %q, want post.created", pub.routingKey)
	}
	if pub.exchange != "forum.events" {
		t.Fatalf("exchange = %q, want forum.events", pub.exchange)
	}
	if db.publishedID != 7 {
		t.Fatalf("published id = %d, want 7", db.publishedID)
	}
	var body map[string]any
	if err := json.Unmarshal(pub.body, &body); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"eventId", "eventType", "aggregateType", "aggregateId", "occurredAt", "payload"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("published body missing envelope key %q: %s", key, pub.body)
		}
	}
	if _, ok := body["EventID"]; ok {
		t.Fatalf("published body uses outbox record fields, want event envelope: %s", pub.body)
	}
}

func TestPublisherUsesAIExchangeForPostTagged(t *testing.T) {
	db := &recordingDBTX{
		pendingRows: []Record{{
			ID:        9,
			EventID:   "evt-9",
			EventType: "post.tagged",
			Payload:   json.RawMessage(`{}`),
		}},
	}
	pub := &recordingPublisher{}
	publisher := NewPublisher(db, pub, Options{BatchSize: 100, MaxRetries: 3, ScanInterval: time.Hour})

	if err := publisher.ProcessOnce(context.Background()); err != nil {
		t.Fatal(err)
	}

	if pub.exchange != "ai.events" {
		t.Fatalf("exchange = %q, want ai.events", pub.exchange)
	}
}

func TestPublisherProcessOnceMarksFailedAfterRetries(t *testing.T) {
	db := &recordingDBTX{
		pendingRows: []Record{{
			ID:         8,
			EventID:    "evt-8",
			EventType:  "post.created",
			Payload:    json.RawMessage(`{}`),
			RetryCount: 2,
		}},
	}
	pub := &recordingPublisher{err: errors.New("broker down")}
	publisher := NewPublisher(db, pub, Options{BatchSize: 100, MaxRetries: 3, ScanInterval: time.Hour})

	if err := publisher.ProcessOnce(context.Background()); err != nil {
		t.Fatal(err)
	}

	if db.failedID != 8 {
		t.Fatalf("failed id = %d, want 8", db.failedID)
	}
}

func TestPublisherStopReturnsAfterStartContextCanceled(t *testing.T) {
	db := &recordingDBTX{}
	pub := &recordingPublisher{}
	publisher := NewPublisher(db, pub, Options{ScanInterval: time.Hour})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- publisher.Start(ctx)
	}()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("publisher did not stop after context cancellation")
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	if err := publisher.Stop(stopCtx); err != nil {
		t.Fatal(err)
	}
}

func TestPublisherShutdownMidPublishLeavesRowPending(t *testing.T) {
	db := &recordingDBTX{
		pendingRows: []Record{{
			ID:        10,
			EventID:   "evt-10",
			EventType: "post.created",
			Payload:   json.RawMessage(`{}`),
		}},
	}
	pub := &blockingPublisher{started: make(chan struct{}), release: make(chan struct{})}
	publisher := NewPublisher(db, pub, Options{BatchSize: 100, MaxRetries: 3, ScanInterval: time.Hour})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = publisher.Start(ctx) }()

	select {
	case <-pub.started:
	case <-time.After(time.Second):
		t.Fatal("publish did not start")
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stopCancel()
	if err := publisher.Stop(stopCtx); err == nil {
		t.Fatal("expected stop timeout while publish is in flight")
	}
	if db.publishedID != 0 || db.failedID != 0 {
		t.Fatalf("row should remain pending, published=%d failed=%d", db.publishedID, db.failedID)
	}
	close(pub.release)
	cancel()
}

type recordingDBTX struct {
	query       string
	args        []any
	pendingRows []Record
	publishedID int64
	failedID    int64
}

func (r *recordingDBTX) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	r.query = query
	r.args = args
	lower := strings.ToLower(query)
	if strings.Contains(lower, "status = 'published'") {
		r.publishedID = args[len(args)-1].(int64)
	}
	if strings.Contains(lower, "status = 'failed'") {
		r.failedID = args[len(args)-1].(int64)
	}
	return fakeResult(1), nil
}

func (r *recordingDBTX) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	panic("unexpected QueryContext")
}

func (r *recordingDBTX) QueryxContext(context.Context, string, ...any) (*sql.Rows, error) {
	panic("unexpected QueryxContext")
}

func (r *recordingDBTX) QueryRowxContext(context.Context, string, ...any) *sql.Row {
	panic("unexpected QueryRowxContext")
}

func (r *recordingDBTX) GetContext(context.Context, interface{}, string, ...interface{}) error {
	panic("unexpected GetContext")
}

func (r *recordingDBTX) SelectContext(_ context.Context, dest interface{}, _ string, _ ...interface{}) error {
	records := dest.(*[]Record)
	*records = append(*records, r.pendingRows...)
	return nil
}

type fakeResult int64

func (f fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (f fakeResult) RowsAffected() (int64, error) { return int64(f), nil }

var _ driver.Result = fakeResult(0)

type recordingPublisher struct {
	exchange   string
	routingKey string
	body       []byte
	err        error
}

func (p *recordingPublisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	p.exchange = exchange
	p.routingKey = routingKey
	p.body = append([]byte(nil), body...)
	return p.err
}

type blockingPublisher struct {
	started chan struct{}
	release chan struct{}
}

func (p *blockingPublisher) Publish(context.Context, string, string, []byte) error {
	close(p.started)
	<-p.release
	return nil
}
