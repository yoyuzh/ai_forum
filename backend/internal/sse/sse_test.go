package sse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHubPublishesToPostSubscribersOnly(t *testing.T) {
	hub := NewHub()
	ch, cancel := hub.Subscribe(42)
	defer cancel()
	other, otherCancel := hub.Subscribe(7)
	defer otherCancel()

	if err := hub.Publish(context.Background(), 42, Event{Type: "ai_reply_completed"}); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-ch:
		if got.Type != "ai_reply_completed" {
			t.Fatalf("event = %#v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
	select {
	case got := <-other:
		t.Fatalf("unexpected other post event: %#v", got)
	default:
	}
}

func TestEventsHandlerStreamsPublishedEvent(t *testing.T) {
	hub := NewHub()
	handler := NewEventsHandler(hub)
	req := httptest.NewRequest(http.MethodGet, "/api/posts/42/events", nil)
	req.SetPathValue("postId", "42")
	rec := httptest.NewRecorder()
	done := make(chan struct{})

	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()
	time.Sleep(10 * time.Millisecond)
	_ = hub.Publish(context.Background(), 42, Event{Type: "ai_reply_completed"})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for SSE handler")
	}
	if !strings.Contains(rec.Body.String(), "event: ai_reply_completed") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestStatusHandlerReturnsRunningAndCompletedCounts(t *testing.T) {
	store := statusStore{completed: 1, running: 1, failed: 1}
	handler := NewStatusHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/api/posts/42/ai-status", nil)
	req.SetPathValue("postId", "42")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`"completedCount":1`, `"runningCount":1`, `"failedCount":1`, `"overallStatus":"RUNNING"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %s: %s", want, body)
		}
	}
}

func TestStatusHandlerReturnsFailedWhenOnlyFailuresExist(t *testing.T) {
	store := statusStore{failed: 1}
	handler := NewStatusHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/api/posts/42/ai-status", nil)
	req.SetPathValue("postId", "42")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	for _, want := range []string{`"failedCount":1`, `"overallStatus":"FAILED"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %s: %s", want, body)
		}
	}
}

type statusStore struct {
	completed int
	running   int
	failed    int
}

func (s statusStore) AIStatus(context.Context, int64) (Status, error) {
	return Status{CompletedCount: s.completed, RunningCount: s.running, FailedCount: s.failed}, nil
}
