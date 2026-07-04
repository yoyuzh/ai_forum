package internalapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientPostsInternalEventWithToken(t *testing.T) {
	var gotPath, gotToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotToken = r.Header.Get("X-Internal-Token")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	client := NewClient(server.URL, "secret", server.Client())

	if err := client.Notify(context.Background(), 42, Event{Type: "ai_reply_completed"}); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/internal/posts/42/events" || gotToken != "secret" {
		t.Fatalf("request = %s token %q", gotPath, gotToken)
	}
}
