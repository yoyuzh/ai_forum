package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenManagerIssuesAndValidatesJWT(t *testing.T) {
	tm := NewTokenManager("test-secret", time.Hour)

	token, err := tm.Issue(Subject{UserID: 7, Username: "alice", Role: "USER"})
	if err != nil {
		t.Fatal(err)
	}
	sub, err := tm.Validate(token)
	if err != nil {
		t.Fatal(err)
	}
	if sub.UserID != 7 || sub.Username != "alice" || sub.Role != "USER" {
		t.Fatalf("subject = %#v", sub)
	}
}

func TestTokenManagerRejectsExpiredJWT(t *testing.T) {
	tm := NewTokenManager("test-secret", -time.Hour)

	token, err := tm.Issue(Subject{UserID: 7, Username: "alice", Role: "USER"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tm.Validate(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestMiddlewarePopulatesSubjectAndRejectsExpiredToken(t *testing.T) {
	tm := NewTokenManager("test-secret", time.Hour)
	token, err := tm.Issue(Subject{UserID: 7, Username: "alice", Role: "USER"})
	if err != nil {
		t.Fatal(err)
	}
	var got Subject
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = SubjectFromContext(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	tm.Middleware(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got.UserID != 7 {
		t.Fatalf("subject = %#v", got)
	}

	expiredTM := NewTokenManager("test-secret", -time.Hour)
	expired, err := expiredTM.Issue(Subject{UserID: 7, Username: "alice", Role: "USER"})
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	rec = httptest.NewRecorder()
	tm.Middleware(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expired status = %d, want 401", rec.Code)
	}
}
