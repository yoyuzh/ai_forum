// Package router owns HTTP route registration and middleware composition.
package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Dependency is one readiness check.
type Dependency struct {
	Name  string
	Check func(context.Context) error
}

type BusinessRoutes struct {
	Register              http.Handler
	Login                 http.Handler
	Profile               http.Handler
	ListPosts             http.Handler
	GetPost               http.Handler
	CreatePost            http.Handler
	UpdatePost            http.Handler
	DeletePost            http.Handler
	CreateComment         http.Handler
	AdminUpdatePostStatus http.Handler
}

// New builds the api-server router.
func New(deps []Dependency, internal http.Handler) http.Handler {
	return NewWithBusinessRoutes(deps, internal, BusinessRoutes{})
}

func NewWithBusinessRoutes(deps []Dependency, internal http.Handler, business BusinessRoutes) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /readyz", readyz(deps))
	if business.ListPosts != nil {
		mux.Handle("GET /api/posts", business.ListPosts)
	} else {
		mux.HandleFunc("GET /api/posts", func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode([]any{})
		})
	}
	if business.GetPost != nil {
		mux.Handle("GET /api/posts/{postId}", business.GetPost)
	}
	if business.Register != nil {
		mux.Handle("POST /api/register", business.Register)
	}
	if business.Login != nil {
		mux.Handle("POST /api/login", business.Login)
	}
	if business.Profile != nil {
		mux.Handle("GET /api/me", business.Profile)
	}
	if business.CreatePost != nil {
		mux.Handle("POST /api/posts", business.CreatePost)
	}
	if business.UpdatePost != nil {
		mux.Handle("PATCH /api/posts/{postId}", business.UpdatePost)
	}
	if business.DeletePost != nil {
		mux.Handle("DELETE /api/posts/{postId}", business.DeletePost)
	}
	if business.CreateComment != nil {
		mux.Handle("POST /api/posts/{postId}/comments", business.CreateComment)
	}
	if business.AdminUpdatePostStatus != nil {
		mux.Handle("PATCH /api/admin/posts/{postId}/status", business.AdminUpdatePostStatus)
	}
	if internal != nil {
		mux.Handle("/internal/", internal)
	}
	return mux
}

func readyz(deps []Dependency) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		var failed []string
		for _, dep := range deps {
			if dep.Check == nil {
				continue
			}
			if err := dep.Check(ctx); err != nil {
				failed = append(failed, dep.Name)
			}
		}
		if len(failed) > 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string][]string{"failed": failed})
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	}
}
