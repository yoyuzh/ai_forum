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
		UpdateProfile: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ProfileStats: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
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
		ListComments: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		LikePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		UnlikePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		FavoritePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		UnfavoritePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		ListNotifications: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		UnreadNotifications: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		MarkNotificationRead: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		MarkAllNotificationsRead: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		PostEvents: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		AIStatus: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ListAgents: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ListAgentChatConversations: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		GetAgentChatConversation: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		StreamAgentChatMessage: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		DeleteAgentChatConversation: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		RetryAgentChatMessage: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		ListAITasks: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ListDecisionLogs: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ListPostDecisionLogs: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ListPostAITasks: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ListAIActivities: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		SearchPosts: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		AdminUpdatePostStatus: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		AdminDashboardStats: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		AdminListDecisionLogs: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		AdminRetryTask: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
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
		{http.MethodPatch, "/api/me", http.StatusOK},
		{http.MethodGet, "/api/me/stats", http.StatusOK},
		{http.MethodGet, "/api/posts", http.StatusOK},
		{http.MethodGet, "/api/posts/42", http.StatusOK},
		{http.MethodPost, "/api/posts", http.StatusCreated},
		{http.MethodPatch, "/api/posts/42", http.StatusOK},
		{http.MethodDelete, "/api/posts/42", http.StatusNoContent},
		{http.MethodGet, "/api/posts/42/comments", http.StatusOK},
		{http.MethodPost, "/api/posts/42/comments", http.StatusCreated},
		{http.MethodPost, "/api/posts/42/like", http.StatusNoContent},
		{http.MethodDelete, "/api/posts/42/like", http.StatusNoContent},
		{http.MethodPost, "/api/posts/42/favorite", http.StatusNoContent},
		{http.MethodDelete, "/api/posts/42/favorite", http.StatusNoContent},
		{http.MethodGet, "/api/notifications", http.StatusOK},
		{http.MethodGet, "/api/notifications/unread-count", http.StatusOK},
		{http.MethodPut, "/api/notifications/9/read", http.StatusNoContent},
		{http.MethodPut, "/api/notifications/read-all", http.StatusNoContent},
		{http.MethodGet, "/api/posts/42/events", http.StatusOK},
		{http.MethodGet, "/api/posts/42/ai-status", http.StatusOK},
		{http.MethodGet, "/api/agents", http.StatusOK},
		{http.MethodGet, "/api/ai-chat/conversations", http.StatusOK},
		{http.MethodGet, "/api/ai-chat/conversations/1001/messages", http.StatusOK},
		{http.MethodPost, "/api/ai-chat/messages/stream", http.StatusCreated},
		{http.MethodDelete, "/api/ai-chat/conversations/1001", http.StatusOK},
		{http.MethodPost, "/api/ai-chat/messages/2002/retry", http.StatusCreated},
		{http.MethodGet, "/api/ai-tasks", http.StatusOK},
		{http.MethodGet, "/api/decision-logs", http.StatusOK},
		{http.MethodGet, "/api/posts/42/decision-logs", http.StatusOK},
		{http.MethodGet, "/api/posts/42/ai-tasks", http.StatusOK},
		{http.MethodGet, "/api/ai-activity", http.StatusOK},
		{http.MethodGet, "/api/search/posts?q=hello", http.StatusOK},
		{http.MethodPatch, "/api/admin/posts/42/status", http.StatusNoContent},
		{http.MethodGet, "/api/admin/dashboard/stats", http.StatusOK},
		{http.MethodGet, "/api/admin/decision-logs", http.StatusOK},
		{http.MethodPost, "/api/admin/ai-tasks/42/retry", http.StatusForbidden},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != tc.status {
			t.Fatalf("%s %s status = %d, want %d", tc.method, tc.path, rec.Code, tc.status)
		}
	}
}
