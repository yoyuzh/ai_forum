package user

import (
	"encoding/json"
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

func TestProfileUpdateAndStatsHandlersUseCurrentSubject(t *testing.T) {
	svc := NewService(newMemoryRepository())
	u, err := svc.Register(httptest.NewRequest(http.MethodGet, "/", nil).Context(), RegisterInput{Username: "alice", Password: "secret123", DisplayName: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(`{"nickname":"Alice B"}`))
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: u.ID, Username: "alice", Role: "USER"}))
	rec := httptest.NewRecorder()
	h.UpdateProfile(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["display_name"] != "Alice B" {
		t.Fatalf("display_name = %v, want Alice B", body["display_name"])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/me/stats", nil)
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: u.ID, Username: "alice", Role: "USER"}))
	rec = httptest.NewRecorder()
	h.Stats(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("stats status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
