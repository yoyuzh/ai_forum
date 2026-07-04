// Package admin exposes backend-authoritative admin REST handlers.
package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/auth"
	"ai-forum/backend/internal/rbac"
)

type User struct {
	ID        int64  `json:"id" db:"id"`
	Username  string `json:"username" db:"username"`
	Role      string `json:"role" db:"role"`
	Status    string `json:"status" db:"status"`
	PostCount int64  `json:"postCount" db:"post_count"`
	CreatedAt string `json:"createdAt" db:"created_at"`
}

type Post struct {
	ID               int64  `json:"id" db:"id"`
	Title            string `json:"title" db:"title"`
	Author           string `json:"author" db:"author"`
	Status           string `json:"status" db:"status"`
	ViewCount        int64  `json:"viewCount" db:"view_count"`
	CommentCount     int64  `json:"commentCount" db:"comment_count"`
	AIResponsesCount int64  `json:"aiResponsesCount" db:"ai_reply_count"`
	CreatedAt        string `json:"createdAt" db:"created_at"`
}

type Comment struct {
	ID        int64  `json:"id" db:"id"`
	PostID    int64  `json:"postId" db:"post_id"`
	Author    string `json:"author" db:"author"`
	Type      string `json:"type" db:"comment_type"`
	Content   string `json:"content" db:"content"`
	CreatedAt string `json:"createdAt" db:"created_at"`
}

type Agent struct {
	ID                 int64   `json:"id" db:"id"`
	Name               string  `json:"name" db:"name"`
	SystemPrompt       string  `json:"systemPrompt" db:"system_prompt"`
	ReplyThreshold     float64 `json:"replyThreshold" db:"reply_threshold"`
	ActivityLevel      float64 `json:"activityLevel" db:"activity_level"`
	AllowAutoReply     bool    `json:"allowAutoReply" db:"allow_auto_reply"`
	AllowMentionReply  bool    `json:"allowMentionReply" db:"allow_mention"`
	AllowFollowupReply bool    `json:"allowFollowupReply" db:"allow_followup"`
	Active             bool    `json:"active" db:"enabled"`
	Fallback           bool    `json:"fallback" db:"is_fallback"`
	ReplyCount         int64   `json:"replyCount" db:"reply_count"`
	CreatedAt          string  `json:"createdAt" db:"created_at"`
}

type Task struct {
	ID              int64   `json:"id" db:"id"`
	PostID          int64   `json:"postId" db:"post_id"`
	ParentCommentID *int64  `json:"parentCommentId" db:"parent_comment_id"`
	CommentID       *int64  `json:"commentId" db:"comment_id"`
	AIAgentID       int64   `json:"aiAgentId" db:"ai_agent_id"`
	AIAgentName     string  `json:"aiAgentName" db:"ai_agent_name"`
	TriggerType     string  `json:"triggerType" db:"trigger_type"`
	Status          string  `json:"status" db:"status"`
	RetryCount      int64   `json:"retryCount" db:"attempt_count"`
	ErrorMessage    *string `json:"errorMessage" db:"last_error"`
	CreatedAt       string  `json:"createdAt" db:"created_at"`
	UpdatedAt       string  `json:"updatedAt" db:"updated_at"`
}

type DecisionLog struct {
	ID               int64    `json:"id" db:"id"`
	PostID           int64    `json:"postId" db:"post_id"`
	CommentID        *int64   `json:"commentId" db:"comment_id"`
	AIAgentID        int64    `json:"aiAgentId" db:"ai_agent_id"`
	AIAgentName      string   `json:"aiAgentName" db:"ai_agent_name"`
	TriggerType      string   `json:"triggerType" db:"trigger_type"`
	WillingnessScore float64  `json:"willingnessScore" db:"willingness_score"`
	ThresholdValue   float64  `json:"thresholdValue" db:"threshold_value"`
	Decision         string   `json:"decision" db:"decision"`
	Reason           string   `json:"reason" db:"reason"`
	Fallback         bool     `json:"fallback" db:"fallback"`
	HitTags          []string `json:"hitTags" db:"-"`
	TaskID           *int64   `json:"taskId" db:"task_id"`
	CommentLink      *int64   `json:"commentLink" db:"reply_comment_id"`
	CreatedAt        string   `json:"createdAt" db:"created_at"`
}

type Tag struct {
	ID        int64  `json:"id" db:"id"`
	PostID    int64  `json:"postId" db:"post_id"`
	Type      string `json:"type" db:"tag_type"`
	Name      string `json:"name" db:"tag_name"`
	CreatedAt string `json:"createdAt" db:"created_at"`
}

type Preference struct {
	ID        int64   `json:"id" db:"id"`
	AgentID   int64   `json:"agentId" db:"ai_agent_id"`
	TagType   string  `json:"tagType" db:"tag_type"`
	TagName   string  `json:"tagName" db:"tag_name"`
	Weight    float64 `json:"weight" db:"weight"`
	CreatedAt string  `json:"createdAt" db:"created_at"`
}

type PublicAgent struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Avatar      string   `json:"avatar"`
	Icon        string   `json:"icon"`
	Description string   `json:"description"`
	Traits      []string `json:"traits"`
	Specialties []string `json:"specialties"`
	Active      bool     `json:"active"`
}

type PublicActivity struct {
	ID           int64  `json:"id"`
	AgentName    string `json:"agentName"`
	AgentAvatar  string `json:"agentAvatar"`
	Action       string `json:"action"`
	Target       string `json:"target"`
	TargetID     int64  `json:"targetId"`
	RelativeTime string `json:"relativeTime"`
}

type DashboardStats struct {
	TotalUsers   int64 `json:"totalUsers" db:"total_users"`
	TotalPosts   int64 `json:"totalPosts" db:"total_posts"`
	AIReplies    int64 `json:"aiReplies" db:"ai_replies"`
	TodayAITasks int64 `json:"todayAiTasks" db:"today_ai_tasks"`
	FailedTasks  int64 `json:"failedTasks" db:"failed_tasks"`
}

type TrendPoint struct {
	Label string `json:"label" db:"label"`
	Value int64  `json:"value" db:"value"`
}

type TaskStatusBreakdown struct {
	Success int64 `json:"success"`
	Running int64 `json:"running"`
	Failed  int64 `json:"failed"`
}

type ServiceStatus struct {
	Name    string `json:"name"`
	Metric  string `json:"metric"`
	Healthy bool   `json:"healthy"`
}

type RecentPost struct {
	ID           int64  `json:"id" db:"id"`
	Title        string `json:"title" db:"title"`
	Author       string `json:"author" db:"author"`
	RelativeTime string `json:"relativeTime" db:"relative_time"`
	Status       string `json:"status" db:"status"`
}

type RecentTask struct {
	ID     int64  `json:"id" db:"id"`
	Label  string `json:"label" db:"label"`
	Icon   string `json:"icon"`
	Status string `json:"status" db:"status"`
}

type DecisionTimelineEntry struct {
	Time    string `json:"time" db:"time"`
	Message string `json:"message" db:"message"`
}

type AgentUpdate struct {
	ReplyThreshold     *float64 `json:"replyThreshold"`
	ActivityLevel      *float64 `json:"activityLevel"`
	SystemPrompt       *string  `json:"systemPrompt"`
	AllowAutoReply     *bool    `json:"allowAutoReply"`
	AllowMentionReply  *bool    `json:"allowMentionReply"`
	AllowFollowupReply *bool    `json:"allowFollowupReply"`
	Active             *bool    `json:"active"`
}

type Store interface {
	DashboardStats(context.Context) (DashboardStats, error)
	WeeklyTrend(context.Context) ([]TrendPoint, error)
	TaskStatusBreakdown(context.Context) (TaskStatusBreakdown, error)
	Services(context.Context) ([]ServiceStatus, error)
	RecentPosts(context.Context) ([]RecentPost, error)
	RecentTasks(context.Context) ([]RecentTask, error)
	DecisionTimeline(context.Context) ([]DecisionTimelineEntry, error)
	ListUsers(context.Context) ([]User, error)
	ListPosts(context.Context) ([]Post, error)
	ListComments(context.Context) ([]Comment, error)
	ListAgents(context.Context) ([]Agent, error)
	UpdateAgent(context.Context, int64, AgentUpdate) (Agent, error)
	ListTasks(context.Context) ([]Task, error)
	RetryTask(context.Context, int64) (Task, error)
	TerminateTask(context.Context, int64) (Task, error)
	MarkTaskProcessed(context.Context, int64) (Task, error)
	ListDecisionLogs(context.Context) ([]DecisionLog, error)
	ListTags(context.Context) ([]Tag, error)
	ListPreferences(context.Context) ([]Preference, error)
}

type Handler struct {
	store Store
	authz *rbac.Authorizer
}

func NewHandler(store Store, authz *rbac.Authorizer) *Handler {
	return &Handler{store: store, authz: authz}
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.ListUsers(ctx) })
}
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.ListPosts(ctx) })
}
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.ListComments(ctx) })
}
func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.ListAgents(ctx) })
}
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.ListTasks(ctx) })
}
func (h *Handler) ListDecisionLogs(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.ListDecisionLogs(ctx) })
}
func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.ListTags(ctx) })
}
func (h *Handler) ListPreferences(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.ListPreferences(ctx) })
}

func (h *Handler) ListPublicAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.store.ListAgents(r.Context())
	if err != nil {
		http.Error(w, "list agents", http.StatusInternalServerError)
		return
	}
	out := make([]PublicAgent, 0, len(agents))
	for _, agent := range agents {
		if !agent.Active {
			continue
		}
		out = append(out, publicAgent(agent))
	}
	writeJSON(w, out)
}

func (h *Handler) ListPostDecisionLogs(w http.ResponseWriter, r *http.Request) {
	postID, ok := pathInt64(w, r, "postId")
	if !ok {
		return
	}
	logs, err := h.store.ListDecisionLogs(r.Context())
	if err != nil {
		http.Error(w, "list decision logs", http.StatusInternalServerError)
		return
	}
	out := make([]DecisionLog, 0, len(logs))
	for _, log := range logs {
		if log.PostID == postID {
			out = append(out, log)
		}
	}
	writeJSON(w, out)
}

func (h *Handler) ListPublicDecisionLogs(w http.ResponseWriter, r *http.Request) {
	h.ListDecisionLogs(w, r)
}

func (h *Handler) ListPostTasks(w http.ResponseWriter, r *http.Request) {
	postID, ok := pathInt64(w, r, "postId")
	if !ok {
		return
	}
	tasks, err := h.store.ListTasks(r.Context())
	if err != nil {
		http.Error(w, "list ai tasks", http.StatusInternalServerError)
		return
	}
	out := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		if task.PostID == postID {
			out = append(out, task)
		}
	}
	writeJSON(w, out)
}

func (h *Handler) ListPublicTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.store.ListTasks(r.Context())
	if err != nil {
		http.Error(w, "list ai tasks", http.StatusInternalServerError)
		return
	}
	writeJSON(w, tasks)
}

func (h *Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	logs, err := h.store.ListDecisionLogs(r.Context())
	if err != nil {
		http.Error(w, "list ai activity", http.StatusInternalServerError)
		return
	}
	out := make([]PublicActivity, 0, len(logs))
	for _, log := range logs {
		out = append(out, PublicActivity{
			ID:           log.ID,
			AgentName:    log.AIAgentName,
			AgentAvatar:  "",
			Action:       log.Decision,
			Target:       "post",
			TargetID:     log.PostID,
			RelativeTime: log.CreatedAt,
		})
		if len(out) == 20 {
			break
		}
	}
	writeJSON(w, out)
}

func (h *Handler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.DashboardStats(ctx) })
}
func (h *Handler) WeeklyTrend(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.WeeklyTrend(ctx) })
}
func (h *Handler) TaskStatusBreakdown(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.TaskStatusBreakdown(ctx) })
}
func (h *Handler) Services(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.Services(ctx) })
}
func (h *Handler) RecentPosts(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.RecentPosts(ctx) })
}
func (h *Handler) RecentTasks(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.RecentTasks(ctx) })
}
func (h *Handler) DecisionTimeline(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, func(ctx context.Context) (any, error) { return h.store.DecisionTimeline(ctx) })
}

func (h *Handler) Permissions(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	all := []string{"post:delete-any", "user:ban", "ai_task:retry", "ai_agent:update", "decision_log:read"}
	var allowed []string
	for _, p := range all {
		obj, act, ok := cutPermission(p)
		if !ok {
			continue
		}
		pass, err := h.authz.Enforce(sub.Role, obj, act)
		if err != nil {
			http.Error(w, "authorize", http.StatusInternalServerError)
			return
		}
		if pass {
			allowed = append(allowed, p)
		}
	}
	writeJSON(w, map[string]any{"role": sub.Role, "permissions": allowed})
}

func (h *Handler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "ai_agent", "update") {
		return
	}
	id, ok := pathInt64(w, r, "agentId")
	if !ok {
		return
	}
	var update AgentUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	agent, err := h.store.UpdateAgent(r.Context(), id, update)
	if err != nil {
		http.Error(w, "update agent", http.StatusBadRequest)
		return
	}
	writeJSON(w, agent)
}

func (h *Handler) RetryTask(w http.ResponseWriter, r *http.Request) {
	h.taskAction(w, r, h.store.RetryTask)
}

func (h *Handler) TerminateTask(w http.ResponseWriter, r *http.Request) {
	h.taskAction(w, r, h.store.TerminateTask)
}

func (h *Handler) MarkTaskProcessed(w http.ResponseWriter, r *http.Request) {
	h.taskAction(w, r, h.store.MarkTaskProcessed)
}

func (h *Handler) taskAction(w http.ResponseWriter, r *http.Request, fn func(context.Context, int64) (Task, error)) {
	if !h.require(w, r, "ai_task", "retry") {
		return
	}
	id, ok := pathInt64(w, r, "taskId")
	if !ok {
		return
	}
	task, err := fn(r.Context(), id)
	if err != nil {
		http.Error(w, "update task", http.StatusBadRequest)
		return
	}
	writeJSON(w, task)
}

func (h *Handler) require(w http.ResponseWriter, r *http.Request, obj, act string) bool {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	allowed, err := h.authz.Enforce(sub.Role, obj, act)
	if err != nil {
		http.Error(w, "authorize", http.StatusInternalServerError)
		return false
	}
	if !allowed {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, fn func(context.Context) (any, error)) {
	rows, err := fn(r.Context())
	if err != nil {
		http.Error(w, "list resource", http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sqlx.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) DashboardStats(ctx context.Context) (DashboardStats, error) {
	var row DashboardStats
	err := s.db.GetContext(ctx, &row, `
		SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL) AS total_posts,
			(SELECT COALESCE(SUM(ai_reply_count), 0) FROM posts WHERE deleted_at IS NULL) AS ai_replies,
			(SELECT COUNT(*) FROM ai_reply_tasks WHERE DATE(created_at) = UTC_DATE()) AS today_ai_tasks,
			(SELECT COUNT(*) FROM ai_reply_tasks WHERE status = 'FAILED') AS failed_tasks`)
	return row, err
}

func (s *SQLStore) WeeklyTrend(ctx context.Context) ([]TrendPoint, error) {
	var rows []TrendPoint
	err := s.db.SelectContext(ctx, &rows, `
		SELECT DATE_FORMAT(created_at, '%m-%d') AS label, COUNT(*) AS value
		FROM posts
		WHERE deleted_at IS NULL AND created_at >= UTC_TIMESTAMP() - INTERVAL 6 DAY
		GROUP BY DATE(created_at), DATE_FORMAT(created_at, '%m-%d')
		ORDER BY DATE(created_at)`)
	if rows == nil {
		rows = []TrendPoint{}
	}
	return rows, err
}

func (s *SQLStore) TaskStatusBreakdown(ctx context.Context) (TaskStatusBreakdown, error) {
	var rows []struct {
		Status string `db:"status"`
		Count  int64  `db:"count"`
	}
	if err := s.db.SelectContext(ctx, &rows, `SELECT status, COUNT(*) AS count FROM ai_reply_tasks GROUP BY status`); err != nil {
		return TaskStatusBreakdown{}, err
	}
	var out TaskStatusBreakdown
	for _, row := range rows {
		switch row.Status {
		case "COMPLETED":
			out.Success = row.Count
		case "PROCESSING", "PENDING":
			out.Running += row.Count
		case "FAILED":
			out.Failed = row.Count
		}
	}
	return out, nil
}

func (s *SQLStore) Services(ctx context.Context) ([]ServiceStatus, error) {
	stats, err := s.DashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	return []ServiceStatus{
		{Name: "MySQL", Metric: "Live", Healthy: true},
		{Name: "Posts", Metric: strconv.FormatInt(stats.TotalPosts, 10), Healthy: true},
		{Name: "AI Tasks", Metric: strconv.FormatInt(stats.TodayAITasks, 10), Healthy: stats.FailedTasks == 0},
		{Name: "Decision Logs", Metric: strconv.FormatInt(stats.AIReplies, 10), Healthy: true},
	}, nil
}

func (s *SQLStore) RecentPosts(ctx context.Context) ([]RecentPost, error) {
	var rows []RecentPost
	err := s.db.SelectContext(ctx, &rows, `
		SELECT p.id, p.title, u.username AS author, DATE_FORMAT(p.created_at, '%Y-%m-%dT%TZ') AS relative_time,
		       CASE WHEN p.status = 'NORMAL' THEN 'published' ELSE 'review' END AS status
		FROM posts p JOIN users u ON u.id = p.author_id
		WHERE p.deleted_at IS NULL
		ORDER BY p.id DESC LIMIT 5`)
	if rows == nil {
		rows = []RecentPost{}
	}
	return rows, err
}

func (s *SQLStore) RecentTasks(ctx context.Context) ([]RecentTask, error) {
	var rows []RecentTask
	err := s.db.SelectContext(ctx, &rows, `
		SELECT t.id, CONCAT(a.name, ' #', t.post_id) AS label, t.status
		FROM ai_reply_tasks t JOIN ai_agents a ON a.id = t.ai_agent_id
		ORDER BY t.id DESC LIMIT 5`)
	for i := range rows {
		rows[i].Icon = "smart_toy"
	}
	if rows == nil {
		rows = []RecentTask{}
	}
	return rows, err
}

func (s *SQLStore) DecisionTimeline(ctx context.Context) ([]DecisionTimelineEntry, error) {
	var rows []DecisionTimelineEntry
	err := s.db.SelectContext(ctx, &rows, `
		SELECT DATE_FORMAT(d.created_at, '%H:%i') AS time,
		       CONCAT(a.name, ' ', d.decision, ' post #', d.post_id) AS message
		FROM decision_logs d JOIN ai_agents a ON a.id = d.ai_agent_id
		ORDER BY d.id DESC LIMIT 6`)
	if rows == nil {
		rows = []DecisionTimelineEntry{}
	}
	return rows, err
}

func (s *SQLStore) ListUsers(ctx context.Context) ([]User, error) {
	var rows []User
	err := s.db.SelectContext(ctx, &rows, `
		SELECT u.id, u.username, u.role, u.status, COUNT(p.id) AS post_count, DATE_FORMAT(u.created_at, '%Y-%m-%dT%TZ') AS created_at
		FROM users u
		LEFT JOIN posts p ON p.author_id = u.id AND p.deleted_at IS NULL
		GROUP BY u.id, u.username, u.role, u.status, u.created_at
		ORDER BY u.id DESC LIMIT 100`)
	return rows, err
}

func (s *SQLStore) ListPosts(ctx context.Context) ([]Post, error) {
	var rows []Post
	err := s.db.SelectContext(ctx, &rows, `
		SELECT p.id, p.title, u.username AS author, p.status, p.view_count, p.comment_count, p.ai_reply_count,
		       DATE_FORMAT(p.created_at, '%Y-%m-%dT%TZ') AS created_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.deleted_at IS NULL
		ORDER BY p.id DESC LIMIT 100`)
	return rows, err
}

func (s *SQLStore) ListComments(ctx context.Context) ([]Comment, error) {
	var rows []Comment
	err := s.db.SelectContext(ctx, &rows, `
		SELECT c.id, c.post_id, COALESCE(u.username, a.name, 'system') AS author, c.comment_type, c.content,
		       DATE_FORMAT(c.created_at, '%Y-%m-%dT%TZ') AS created_at
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		LEFT JOIN ai_agents a ON a.id = c.ai_agent_id
		WHERE c.deleted_at IS NULL
		ORDER BY c.id DESC LIMIT 100`)
	return rows, err
}

func (s *SQLStore) ListAgents(ctx context.Context) ([]Agent, error) {
	var rows []Agent
	err := s.db.SelectContext(ctx, &rows, `
		SELECT a.id, a.name, COALESCE(a.system_prompt, '') AS system_prompt, a.reply_threshold, a.activity_level, a.allow_auto_reply, a.allow_mention,
		       a.allow_followup, a.enabled, a.is_fallback, COUNT(c.id) AS reply_count,
		       DATE_FORMAT(a.created_at, '%Y-%m-%dT%TZ') AS created_at
		FROM ai_agents a
		LEFT JOIN comments c ON c.ai_agent_id = a.id AND c.comment_type = 'AI' AND c.deleted_at IS NULL
		GROUP BY a.id, a.name, COALESCE(a.system_prompt, ''), a.reply_threshold, a.activity_level, a.allow_auto_reply, a.allow_mention,
		         a.allow_followup, a.enabled, a.is_fallback, a.created_at
		ORDER BY a.id`)
	return rows, err
}

func (s *SQLStore) UpdateAgent(ctx context.Context, id int64, update AgentUpdate) (Agent, error) {
	current, err := s.agent(ctx, id)
	if err != nil {
		return Agent{}, err
	}
	if update.ReplyThreshold != nil {
		current.ReplyThreshold = *update.ReplyThreshold
	}
	if update.ActivityLevel != nil {
		current.ActivityLevel = *update.ActivityLevel
	}
	if update.SystemPrompt != nil {
		current.SystemPrompt = *update.SystemPrompt
	}
	if update.AllowAutoReply != nil {
		current.AllowAutoReply = *update.AllowAutoReply
	}
	if update.AllowMentionReply != nil {
		current.AllowMentionReply = *update.AllowMentionReply
	}
	if update.AllowFollowupReply != nil {
		current.AllowFollowupReply = *update.AllowFollowupReply
	}
	if update.Active != nil {
		current.Active = *update.Active
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE ai_agents
		SET reply_threshold=?, activity_level=?, system_prompt=?, allow_auto_reply=?, allow_mention=?, allow_followup=?, enabled=?
		WHERE id=?`, current.ReplyThreshold, current.ActivityLevel, current.SystemPrompt, current.AllowAutoReply, current.AllowMentionReply, current.AllowFollowupReply, current.Active, id)
	if err != nil {
		return Agent{}, err
	}
	return s.agent(ctx, id)
}

func (s *SQLStore) agent(ctx context.Context, id int64) (Agent, error) {
	var row Agent
	err := s.db.GetContext(ctx, &row, `
		SELECT id, name, COALESCE(system_prompt, '') AS system_prompt, reply_threshold, activity_level, allow_auto_reply, allow_mention,
		       allow_followup, enabled, is_fallback, 0 AS reply_count, DATE_FORMAT(created_at, '%Y-%m-%dT%TZ') AS created_at
		FROM ai_agents WHERE id = ?`, id)
	return row, err
}

func (s *SQLStore) ListTasks(ctx context.Context) ([]Task, error) {
	var rows []Task
	err := s.db.SelectContext(ctx, &rows, `
		SELECT t.id, t.post_id, t.parent_comment_id, t.comment_id, t.ai_agent_id, a.name AS ai_agent_name,
		       t.trigger_type, t.status, t.attempt_count, t.last_error,
		       DATE_FORMAT(t.created_at, '%Y-%m-%dT%TZ') AS created_at,
		       DATE_FORMAT(t.updated_at, '%Y-%m-%dT%TZ') AS updated_at
		FROM ai_reply_tasks t
		JOIN ai_agents a ON a.id = t.ai_agent_id
		ORDER BY t.id DESC LIMIT 100`)
	return rows, err
}

func (s *SQLStore) RetryTask(ctx context.Context, id int64) (Task, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE ai_reply_tasks SET status='PENDING', attempt_count=0, last_error=NULL WHERE id=?`, id)
	if err != nil {
		return Task{}, err
	}
	return s.task(ctx, id)
}

func (s *SQLStore) TerminateTask(ctx context.Context, id int64) (Task, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE ai_reply_tasks SET status='FAILED', last_error='terminated by admin' WHERE id=?`, id)
	if err != nil {
		return Task{}, err
	}
	return s.task(ctx, id)
}

func (s *SQLStore) MarkTaskProcessed(ctx context.Context, id int64) (Task, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE ai_reply_tasks SET status='COMPLETED' WHERE id=?`, id)
	if err != nil {
		return Task{}, err
	}
	return s.task(ctx, id)
}

func (s *SQLStore) task(ctx context.Context, id int64) (Task, error) {
	var row Task
	err := s.db.GetContext(ctx, &row, `
		SELECT t.id, t.post_id, t.parent_comment_id, t.comment_id, t.ai_agent_id, a.name AS ai_agent_name,
		       t.trigger_type, t.status, t.attempt_count, t.last_error,
		       DATE_FORMAT(t.created_at, '%Y-%m-%dT%TZ') AS created_at,
		       DATE_FORMAT(t.updated_at, '%Y-%m-%dT%TZ') AS updated_at
		FROM ai_reply_tasks t JOIN ai_agents a ON a.id = t.ai_agent_id WHERE t.id = ?`, id)
	return row, err
}

func (s *SQLStore) ListDecisionLogs(ctx context.Context) ([]DecisionLog, error) {
	var rows []struct {
		DecisionLog
		HitTags sql.NullString `db:"hit_tags"`
	}
	err := s.db.SelectContext(ctx, &rows, `
		SELECT d.id, d.post_id, d.comment_id, d.ai_agent_id, a.name AS ai_agent_name, d.trigger_type,
		       d.willingness_score, d.threshold_value, d.decision, COALESCE(d.reason, '') AS reason,
		       a.is_fallback AS fallback, d.hit_tags, t.id AS task_id, t.comment_id AS reply_comment_id,
		       DATE_FORMAT(d.created_at, '%Y-%m-%dT%TZ') AS created_at
		FROM decision_logs d
		JOIN ai_agents a ON a.id = d.ai_agent_id
		LEFT JOIN ai_reply_tasks t ON t.post_id = d.post_id AND t.ai_agent_id = d.ai_agent_id AND t.trigger_type = d.trigger_type
		ORDER BY d.id DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	out := make([]DecisionLog, 0, len(rows))
	for _, row := range rows {
		log := row.DecisionLog
		log.HitTags = decodeHitTags(row.HitTags.String)
		out = append(out, log)
	}
	return out, nil
}

func (s *SQLStore) ListTags(ctx context.Context) ([]Tag, error) {
	var rows []Tag
	err := s.db.SelectContext(ctx, &rows, `
		SELECT id, post_id, tag_type, tag_name, DATE_FORMAT(created_at, '%Y-%m-%dT%TZ') AS created_at
		FROM post_tags ORDER BY id DESC LIMIT 200`)
	return rows, err
}

func (s *SQLStore) ListPreferences(ctx context.Context) ([]Preference, error) {
	var rows []Preference
	err := s.db.SelectContext(ctx, &rows, `
		SELECT id, ai_agent_id, tag_type, tag_name, weight, DATE_FORMAT(created_at, '%Y-%m-%dT%TZ') AS created_at
		FROM ai_agent_tag_preferences ORDER BY ai_agent_id, id`)
	return rows, err
}

func decodeHitTags(raw string) []string {
	if raw == "" {
		return nil
	}
	var rows []struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil
	}
	tags := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Type == "" {
			tags = append(tags, row.Name)
		} else {
			tags = append(tags, row.Type+":"+row.Name)
		}
	}
	return tags
}

func publicAgent(agent Agent) PublicAgent {
	icon := "smart_toy"
	description := "AI reply decision agent"
	traits := []string{"decision"}
	if agent.Fallback {
		icon = "support_agent"
		description = "Fallback decision agent"
		traits = []string{"fallback"}
	}
	return PublicAgent{
		ID:          agent.ID,
		Name:        agent.Name,
		DisplayName: agent.Name,
		Icon:        icon,
		Description: description,
		Traits:      traits,
		Specialties: []string{},
		Active:      agent.Active,
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func pathInt64(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func cutPermission(p string) (string, string, bool) {
	for i, c := range p {
		if c == ':' {
			return p[:i], p[i+1:], true
		}
	}
	return "", "", false
}
