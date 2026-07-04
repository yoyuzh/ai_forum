// Package sse owns in-memory server-sent-event state for api-server.
package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// Event is intentionally tiny in P3. P7 extends the hub with real subscribers.
type Event struct {
	Type string `json:"type"`
}

// Hub is the dispatch boundary P7 extends with a real in-memory SSE hub.
type Hub interface {
	Publish(context.Context, int64, Event) error
}

// NoopHub is the P3 placeholder; real dispatch is a P7 responsibility.
type NoopHub struct{}

// Publish accepts an event without dispatching.
func (NoopHub) Publish(context.Context, int64, Event) error { return nil }

type InMemoryHub struct {
	mu   sync.RWMutex
	subs map[int64]map[chan Event]struct{}
}

func NewHub() *InMemoryHub {
	return &InMemoryHub{subs: map[int64]map[chan Event]struct{}{}}
}

func (h *InMemoryHub) Subscribe(postID int64) (<-chan Event, func()) {
	ch := make(chan Event, 8)
	h.mu.Lock()
	if h.subs[postID] == nil {
		h.subs[postID] = map[chan Event]struct{}{}
	}
	h.subs[postID][ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.subs[postID], ch)
		if len(h.subs[postID]) == 0 {
			delete(h.subs, postID)
		}
		close(ch)
		h.mu.Unlock()
	}
}

func (h *InMemoryHub) Publish(_ context.Context, postID int64, event Event) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[postID] {
		select {
		case ch <- event:
		default:
		}
	}
	return nil
}

type Status struct {
	CompletedCount int    `json:"completedCount"`
	RunningCount   int    `json:"runningCount"`
	FailedCount    int    `json:"failedCount"`
	RetryableCount int    `json:"retryableCount"`
	OverallStatus  string `json:"overallStatus"`
}

type StatusStore interface {
	AIStatus(context.Context, int64) (Status, error)
}

func NewEventsHandler(hub *InMemoryHub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.ParseInt(r.PathValue("postId"), 10, 64)
		if err != nil || postID <= 0 {
			http.Error(w, "invalid postId", http.StatusBadRequest)
			return
		}
		events, cancel := hub.Subscribe(postID)
		defer cancel()
		w.Header().Set("Content-Type", "text/event-stream")
		select {
		case event := <-events:
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, mustJSON(event))
		case <-r.Context().Done():
		}
	})
}

func NewStatusHandler(store StatusStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.ParseInt(r.PathValue("postId"), 10, 64)
		if err != nil || postID <= 0 {
			http.Error(w, "invalid postId", http.StatusBadRequest)
			return
		}
		status, err := store.AIStatus(r.Context(), postID)
		if err != nil {
			http.Error(w, "ai status", http.StatusInternalServerError)
			return
		}
		if status.RunningCount > 0 {
			status.OverallStatus = "RUNNING"
		} else if status.CompletedCount > 0 {
			status.OverallStatus = "COMPLETED"
		} else if status.FailedCount > 0 {
			status.OverallStatus = "FAILED"
		} else {
			status.OverallStatus = "IDLE"
		}
		_ = json.NewEncoder(w).Encode(status)
	})
}

func mustJSON(v any) string {
	body, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(body)
}
