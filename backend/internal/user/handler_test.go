package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-forum/backend/internal/auth"
)

func TestRegisterHandlerReturnsCreatedAndConflict(t *testing.T) {
	h := NewHandler(NewService(newMemoryRepository()))

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"username":"alice","password":"secret123"}`))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"username":"alice","password":"secret123"}`))
	rec = httptest.NewRecorder()
	h.Register(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", rec.Code, rec.Body.String())
	}
}

func TestProfileHandlerReturnsCurrentUser(t *testing.T) {
	svc := NewService(newMemoryRepository())
	u, err := svc.Register(httptest.NewRequest(http.MethodGet, "/", nil).Context(), RegisterInput{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: u.ID, Username: "alice", Role: "USER"}))
	rec := httptest.NewRecorder()

	h.Profile(rec, req)

	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "secret") || !strings.Contains(rec.Body.String(), "alice") {
		t.Fatalf("status/body = %d/%s", rec.Code, rec.Body.String())
	}
}
