// Package sse owns in-memory server-sent-event state for api-server.
package sse

import "context"

// Event is intentionally tiny in P3. P7 extends the hub with real subscribers.
type Event struct {
	Type string
}

// Hub is the dispatch boundary P7 extends with a real in-memory SSE hub.
type Hub interface {
	Publish(context.Context, int64, Event) error
}

// NoopHub is the P3 placeholder; real dispatch is a P7 responsibility.
type NoopHub struct{}

// Publish accepts an event without dispatching.
func (NoopHub) Publish(context.Context, int64, Event) error { return nil }
