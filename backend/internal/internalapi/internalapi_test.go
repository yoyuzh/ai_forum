package internalapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/logger"
)

func TestHandlerAcceptsValidInternalToken(t *testing.T) {
	var published bool
	h := NewHandler(config.InternalAPIConfig{Token: "secret-token"}, HubFunc(func(context.Context, int64, Event) error {
		published = true
		return nil
	}), testLogger(t, nil))

	req := httptest.NewRequest(http.MethodPost, "/internal/posts/42/events", strings.NewReader(`{"type":"ai_reply_completed"}`))
	req.Header.Set("X-Internal-Token", "secret-token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if !published {
		t.Fatal("expected event to be published to hub")
	}
}

func TestHandlerRejectsMissingAndWrongInternalTokenWithRedactedLog(t *testing.T) {
	token := "secret-token"
	cases := []struct {
		name  string
		token string
	}{
		{name: "missing"},
		{name: "wrong unequal length", token: "bad"},
		{name: "wrong equal length", token: "wrong-token1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			h := NewHandler(config.InternalAPIConfig{Token: token}, NoopHub{}, testLogger(t, &logs))
			req := httptest.NewRequest(http.MethodPost, "/internal/posts/42/events", nil)
			req.Header.Set("User-Agent", "p3-test")
			req.Header.Set("X-Request-ID", "req-1")
			if tc.token != "" {
				req.Header.Set("X-Internal-Token", tc.token)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
			gotLog := logs.String()
			if strings.Contains(gotLog, token) || (tc.token != "" && strings.Contains(gotLog, tc.token)) {
				t.Fatalf("log leaked token: %s", gotLog)
			}
			for _, want := range []string{"request_id", "path", "client_ip", "user_agent", "reason", "***"} {
				if !strings.Contains(gotLog, want) {
					t.Fatalf("log missing %q: %s", want, gotLog)
				}
			}
		})
	}
}

func testLogger(t *testing.T, buf *bytes.Buffer) *logger.Logger {
	t.Helper()
	if buf == nil {
		buf = &bytes.Buffer{}
	}
	l, err := logger.NewWithWriter(config.LogConfig{Level: "info", Encoding: "json"}, buf)
	if err != nil {
		t.Fatal(err)
	}
	return l
}
