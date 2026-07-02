// Package internalapi owns worker-service to api-server internal endpoints.
package internalapi

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/logger"
	"ai-forum/backend/internal/sse"
)

// Event is the internal SSE notification payload. P7 extends the fields.
type Event = sse.Event

// Hub is the dispatch boundary P7 extends with a real in-memory SSE hub.
type Hub = sse.Hub

// HubFunc adapts a function into a Hub.
type HubFunc func(context.Context, int64, Event) error

// Publish calls f.
func (f HubFunc) Publish(ctx context.Context, postID int64, event Event) error {
	return f(ctx, postID, event)
}

// NoopHub accepts events without dispatching until P7 wires real SSE.
type NoopHub = sse.NoopHub

// NewHandler returns the /internal route handler protected by X-Internal-Token.
func NewHandler(cfg config.InternalAPIConfig, hub Hub, log *logger.Logger) http.Handler {
	if hub == nil {
		hub = NoopHub{}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /internal/posts/{postId}/events", func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.ParseInt(r.PathValue("postId"), 10, 64)
		if err != nil || postID <= 0 {
			http.Error(w, "invalid postId", http.StatusBadRequest)
			return
		}
		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid event", http.StatusBadRequest)
			return
		}
		if err := hub.Publish(r.Context(), postID, event); err != nil {
			http.Error(w, "publish event", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	return tokenMiddleware(cfg.Token, log, mux)
}

func tokenMiddleware(want string, log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Internal-Token")
		if !tokenEqual(got, want) {
			if log != nil {
				log.Warn("internal api auth failed",
					logger.FieldRequestID(requestID(r)),
					zap.String("path", r.URL.Path),
					zap.String("client_ip", clientIP(r)),
					zap.String("user_agent", r.UserAgent()),
					zap.String("reason", failureReason(got)),
					logger.FieldInternalAPI(got),
				)
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func tokenEqual(got, want string) bool {
	if got == "" || want == "" {
		return false
	}
	gotHash := sha256.Sum256([]byte(got))
	wantHash := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(gotHash[:], wantHash[:]) == 1
}

func requestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return "-"
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func failureReason(token string) string {
	if strings.TrimSpace(token) == "" {
		return "missing_token"
	}
	return "invalid_token"
}
