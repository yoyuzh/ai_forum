package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthzAlwaysReturnsOK(t *testing.T) {
	h := New(nil, nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestReadyzReturnsOKWhenDependenciesPass(t *testing.T) {
	h := New([]Dependency{
		{Name: "mysql", Check: func(context.Context) error { return nil }},
		{Name: "redis", Check: func(context.Context) error { return nil }},
	}, nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestReadyzReturnsUnavailableWithFailingDependencyName(t *testing.T) {
	h := New([]Dependency{
		{Name: "mysql", Check: func(context.Context) error { return errors.New("dial refused") }},
		{Name: "redis", Check: func(context.Context) error { return nil }},
	}, nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), "mysql") {
		t.Fatalf("body = %q, want failing dependency name", rec.Body.String())
	}
}

func TestPublicPostListDoesNotRequireJWT(t *testing.T) {
	h := New(nil, nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/posts", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want public 200", rec.Code)
	}
}

func TestRouterMountsBusinessRoutes(t *testing.T) {
	h := NewWithBusinessRoutes(nil, nil, BusinessRoutes{
		Register: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) }),
		Login:    http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
		Profile:  http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
		ListPosts: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		GetPost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		CreatePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		UpdatePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		DeletePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		CreateComment: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		AdminUpdatePostStatus: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	})

	for _, tc := range []struct {
		method string
		path   string
		status int
	}{
		{http.MethodPost, "/api/register", http.StatusCreated},
		{http.MethodPost, "/api/login", http.StatusOK},
		{http.MethodGet, "/api/me", http.StatusOK},
		{http.MethodGet, "/api/posts", http.StatusOK},
		{http.MethodGet, "/api/posts/42", http.StatusOK},
		{http.MethodPost, "/api/posts", http.StatusCreated},
		{http.MethodPatch, "/api/posts/42", http.StatusOK},
		{http.MethodDelete, "/api/posts/42", http.StatusNoContent},
		{http.MethodPost, "/api/posts/42/comments", http.StatusCreated},
		{http.MethodPatch, "/api/admin/posts/42/status", http.StatusNoContent},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != tc.status {
			t.Fatalf("%s %s status = %d, want %d", tc.method, tc.path, rec.Code, tc.status)
		}
	}
}
