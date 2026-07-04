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
	Register                    http.Handler
	Login                       http.Handler
	Profile                     http.Handler
	UpdateProfile               http.Handler
	ProfileStats                http.Handler
	HotTags                     http.Handler
	ListPosts                   http.Handler
	GetPost                     http.Handler
	CreatePost                  http.Handler
	UpdatePost                  http.Handler
	DeletePost                  http.Handler
	ListComments                http.Handler
	CreateComment               http.Handler
	LikePost                    http.Handler
	UnlikePost                  http.Handler
	FavoritePost                http.Handler
	UnfavoritePost              http.Handler
	ListNotifications           http.Handler
	UnreadNotifications         http.Handler
	MarkNotificationRead        http.Handler
	MarkAllNotificationsRead    http.Handler
	PostEvents                  http.Handler
	AIStatus                    http.Handler
	RetryAIReplies              http.Handler
	ListAgents                  http.Handler
	ListAgentChatConversations  http.Handler
	GetAgentChatConversation    http.Handler
	StreamAgentChatMessage      http.Handler
	DeleteAgentChatConversation http.Handler
	RetryAgentChatMessage       http.Handler
	ListAITasks                 http.Handler
	ListDecisionLogs            http.Handler
	ListPostDecisionLogs        http.Handler
	ListPostAITasks             http.Handler
	ListAIActivities            http.Handler
	SearchPosts                 http.Handler
	AdminUpdatePostStatus       http.Handler
	AdminDashboardStats         http.Handler
	AdminDashboardTrend         http.Handler
	AdminDashboardBreakdown     http.Handler
	AdminDashboardServices      http.Handler
	AdminDashboardRecentPosts   http.Handler
	AdminDashboardRecentTasks   http.Handler
	AdminDashboardDecisions     http.Handler
	AdminPermissions            http.Handler
	AdminListUsers              http.Handler
	AdminListPosts              http.Handler
	AdminListComments           http.Handler
	AdminListAgents             http.Handler
	AdminUpdateAgent            http.Handler
	AdminListTasks              http.Handler
	AdminRetryTask              http.Handler
	AdminTerminateTask          http.Handler
	AdminMarkTaskProcessed      http.Handler
	AdminListDecisionLogs       http.Handler
	AdminListTags               http.Handler
	AdminListPreferences        http.Handler
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
	if business.HotTags != nil {
		mux.Handle("GET /api/tags/hot", business.HotTags)
	}
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
	if business.UpdateProfile != nil {
		mux.Handle("PATCH /api/me", business.UpdateProfile)
	}
	if business.ProfileStats != nil {
		mux.Handle("GET /api/me/stats", business.ProfileStats)
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
	if business.ListComments != nil {
		mux.Handle("GET /api/posts/{postId}/comments", business.ListComments)
	}
	if business.LikePost != nil {
		mux.Handle("POST /api/posts/{postId}/like", business.LikePost)
	}
	if business.UnlikePost != nil {
		mux.Handle("DELETE /api/posts/{postId}/like", business.UnlikePost)
	}
	if business.FavoritePost != nil {
		mux.Handle("POST /api/posts/{postId}/favorite", business.FavoritePost)
	}
	if business.UnfavoritePost != nil {
		mux.Handle("DELETE /api/posts/{postId}/favorite", business.UnfavoritePost)
	}
	if business.ListNotifications != nil {
		mux.Handle("GET /api/notifications", business.ListNotifications)
	}
	if business.UnreadNotifications != nil {
		mux.Handle("GET /api/notifications/unread-count", business.UnreadNotifications)
	}
	if business.MarkNotificationRead != nil {
		mux.Handle("PUT /api/notifications/{notificationId}/read", business.MarkNotificationRead)
	}
	if business.MarkAllNotificationsRead != nil {
		mux.Handle("PUT /api/notifications/read-all", business.MarkAllNotificationsRead)
	}
	if business.PostEvents != nil {
		mux.Handle("GET /api/posts/{postId}/events", business.PostEvents)
	}
	if business.AIStatus != nil {
		mux.Handle("GET /api/posts/{postId}/ai-status", business.AIStatus)
	}
	if business.RetryAIReplies != nil {
		mux.Handle("POST /api/posts/{postId}/ai-retry", business.RetryAIReplies)
	}
	if business.ListAgents != nil {
		mux.Handle("GET /api/agents", business.ListAgents)
	}
	if business.ListAgentChatConversations != nil {
		mux.Handle("GET /api/ai-chat/conversations", business.ListAgentChatConversations)
	}
	if business.GetAgentChatConversation != nil {
		mux.Handle("GET /api/ai-chat/conversations/{conversationId}/messages", business.GetAgentChatConversation)
	}
	if business.StreamAgentChatMessage != nil {
		mux.Handle("POST /api/ai-chat/messages/stream", business.StreamAgentChatMessage)
	}
	if business.DeleteAgentChatConversation != nil {
		mux.Handle("DELETE /api/ai-chat/conversations/{conversationId}", business.DeleteAgentChatConversation)
	}
	if business.RetryAgentChatMessage != nil {
		mux.Handle("POST /api/ai-chat/messages/{messageId}/retry", business.RetryAgentChatMessage)
	}
	if business.ListAITasks != nil {
		mux.Handle("GET /api/ai-tasks", business.ListAITasks)
	}
	if business.ListDecisionLogs != nil {
		mux.Handle("GET /api/decision-logs", business.ListDecisionLogs)
	}
	if business.ListPostDecisionLogs != nil {
		mux.Handle("GET /api/posts/{postId}/decision-logs", business.ListPostDecisionLogs)
	}
	if business.ListPostAITasks != nil {
		mux.Handle("GET /api/posts/{postId}/ai-tasks", business.ListPostAITasks)
	}
	if business.ListAIActivities != nil {
		mux.Handle("GET /api/ai-activity", business.ListAIActivities)
	}
	if business.SearchPosts != nil {
		mux.Handle("GET /api/search/posts", business.SearchPosts)
	}
	if business.AdminUpdatePostStatus != nil {
		mux.Handle("PATCH /api/admin/posts/{postId}/status", business.AdminUpdatePostStatus)
	}
	if business.AdminDashboardStats != nil {
		mux.Handle("GET /api/admin/dashboard/stats", business.AdminDashboardStats)
	}
	if business.AdminDashboardTrend != nil {
		mux.Handle("GET /api/admin/dashboard/weekly-trend", business.AdminDashboardTrend)
	}
	if business.AdminDashboardBreakdown != nil {
		mux.Handle("GET /api/admin/dashboard/task-status-breakdown", business.AdminDashboardBreakdown)
	}
	if business.AdminDashboardServices != nil {
		mux.Handle("GET /api/admin/dashboard/services", business.AdminDashboardServices)
	}
	if business.AdminDashboardRecentPosts != nil {
		mux.Handle("GET /api/admin/dashboard/recent-posts", business.AdminDashboardRecentPosts)
	}
	if business.AdminDashboardRecentTasks != nil {
		mux.Handle("GET /api/admin/dashboard/recent-tasks", business.AdminDashboardRecentTasks)
	}
	if business.AdminDashboardDecisions != nil {
		mux.Handle("GET /api/admin/dashboard/decision-timeline", business.AdminDashboardDecisions)
	}
	if business.AdminPermissions != nil {
		mux.Handle("GET /api/admin/permissions", business.AdminPermissions)
	}
	if business.AdminListUsers != nil {
		mux.Handle("GET /api/admin/users", business.AdminListUsers)
	}
	if business.AdminListPosts != nil {
		mux.Handle("GET /api/admin/posts", business.AdminListPosts)
	}
	if business.AdminListComments != nil {
		mux.Handle("GET /api/admin/comments", business.AdminListComments)
	}
	if business.AdminListAgents != nil {
		mux.Handle("GET /api/admin/ai-agents", business.AdminListAgents)
	}
	if business.AdminUpdateAgent != nil {
		mux.Handle("PATCH /api/admin/ai-agents/{agentId}", business.AdminUpdateAgent)
	}
	if business.AdminListTasks != nil {
		mux.Handle("GET /api/admin/ai-tasks", business.AdminListTasks)
	}
	if business.AdminRetryTask != nil {
		mux.Handle("POST /api/admin/ai-tasks/{taskId}/retry", business.AdminRetryTask)
	}
	if business.AdminTerminateTask != nil {
		mux.Handle("POST /api/admin/ai-tasks/{taskId}/terminate", business.AdminTerminateTask)
	}
	if business.AdminMarkTaskProcessed != nil {
		mux.Handle("POST /api/admin/ai-tasks/{taskId}/mark-processed", business.AdminMarkTaskProcessed)
	}
	if business.AdminListDecisionLogs != nil {
		mux.Handle("GET /api/admin/decision-logs", business.AdminListDecisionLogs)
	}
	if business.AdminListTags != nil {
		mux.Handle("GET /api/admin/tags", business.AdminListTags)
	}
	if business.AdminListPreferences != nil {
		mux.Handle("GET /api/admin/preferences", business.AdminListPreferences)
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
