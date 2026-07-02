package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoginHandlerReturnsTokenAndRejectsBadPassword(t *testing.T) {
	h := NewHandler(staticAuthenticator{}, NewTokenManager("secret", time.Hour))

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"alice","password":"secret123"}`))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "token") {
		t.Fatalf("status/body = %d/%s, want token", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"alice","password":"badpass123"}`))
	rec = httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

type staticAuthenticator struct{}

func (staticAuthenticator) Authenticate(_ context.Context, username, password string) (Subject, error) {
	if username == "alice" && password == "secret123" {
		return Subject{UserID: 7, Username: "alice", Role: "USER"}, nil
	}
	return Subject{}, ErrInvalidCredentials
}
