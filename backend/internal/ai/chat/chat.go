// Package chat owns one-to-one user conversations with AI agents.
package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/ai/modelclient"
	"ai-forum/backend/internal/auth"
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

var (
	ErrAgentNotFound = errors.New("agent not found")
	ErrEmptyMessage  = errors.New("empty message")
	ErrModelFailure  = errors.New("model failure")
)

type Agent struct {
	ID           int64  `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	SystemPrompt string `json:"systemPrompt" db:"system_prompt"`
}

type Session struct {
	ID        int64  `json:"id" db:"id"`
	UserID    int64  `json:"userId" db:"user_id"`
	AIAgentID int64  `json:"aiAgentId" db:"ai_agent_id"`
	Title     string `json:"title" db:"title"`
	CreatedAt string `json:"createdAt" db:"created_at"`
	UpdatedAt string `json:"updatedAt" db:"updated_at"`
}

type SessionSummary struct {
	Session      Session `json:"session"`
	Agent        Agent   `json:"agent"`
	LastMessage  string  `json:"lastMessage"`
	MessageCount int64   `json:"messageCount"`
}

type Message struct {
	ID           int64   `json:"id" db:"id"`
	SessionID    int64   `json:"sessionId" db:"session_id"`
	Role         string  `json:"role" db:"role"`
	Content      string  `json:"content" db:"content"`
	ErrorMessage *string `json:"errorMessage,omitempty" db:"error_message"`
	CreatedAt    string  `json:"createdAt" db:"created_at"`
}

type NewMessage struct {
	SessionID int64
	Role      string
	Content   string
}

type Store interface {
	ListSessions(context.Context, int64) ([]SessionSummary, error)
	GetAgent(context.Context, int64) (Agent, error)
	CreateSession(context.Context, int64, int64, string) (Session, error)
	GetLatestOrCreateSession(context.Context, int64, int64, string) (Session, error)
	GetSession(context.Context, int64, int64, int64) (Session, error)
	UpdateSessionTitle(context.Context, int64, string) (Session, error)
	ListMessages(context.Context, int64) ([]Message, error)
	CreateMessage(context.Context, NewMessage) (Message, error)
}

type Service struct {
	store Store
	model modelclient.Client
}

func NewService(store Store, model modelclient.Client) *Service {
	return &Service{store: store, model: model}
}

type SessionResponse struct {
	Session  Session   `json:"session"`
	Agent    Agent     `json:"agent"`
	Messages []Message `json:"messages"`
}

type SendResponse struct {
	Session          Session  `json:"session"`
	UserMessage      Message  `json:"userMessage"`
	AssistantMessage *Message `json:"assistantMessage,omitempty"`
}

func (s *Service) List(ctx context.Context, userID int64) ([]SessionSummary, error) {
	return s.store.ListSessions(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID, agentID int64) (SessionResponse, error) {
	agent, err := s.store.GetAgent(ctx, agentID)
	if err != nil {
		return SessionResponse{}, err
	}
	session, err := s.store.CreateSession(ctx, userID, agentID, agent.Name)
	if err != nil {
		return SessionResponse{}, err
	}
	return SessionResponse{Session: session, Agent: agent, Messages: []Message{}}, nil
}

func (s *Service) Get(ctx context.Context, userID, agentID, sessionID int64) (SessionResponse, error) {
	agent, err := s.store.GetAgent(ctx, agentID)
	if err != nil {
		return SessionResponse{}, err
	}
	session, err := s.session(ctx, userID, agentID, sessionID, agent.Name)
	if err != nil {
		return SessionResponse{}, err
	}
	messages, err := s.store.ListMessages(ctx, session.ID)
	if err != nil {
		return SessionResponse{}, err
	}
	return SessionResponse{Session: session, Agent: agent, Messages: orderMessages(messages)}, nil
}

func (s *Service) Send(ctx context.Context, userID, agentID, sessionID int64, content string) (SendResponse, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return SendResponse{}, ErrEmptyMessage
	}
	current, err := s.Get(ctx, userID, agentID, sessionID)
	if err != nil {
		return SendResponse{}, err
	}
	userMsg, err := s.store.CreateMessage(ctx, NewMessage{SessionID: current.Session.ID, Role: RoleUser, Content: content})
	if err != nil {
		return SendResponse{}, err
	}
	session := current.Session
	if len(current.Messages) == 0 {
		updated, err := s.store.UpdateSessionTitle(ctx, session.ID, makeSessionTitle(content))
		if err == nil {
			session = updated
		}
	}
	resp := SendResponse{Session: session, UserMessage: userMsg}
	reply, err := s.model.Generate(ctx, modelclient.Request{
		SystemPrompt: current.Agent.SystemPrompt,
		Prompt:       buildPrompt(current.Agent, current.Messages, content),
		TaskType:     "ai_chat",
		AIAgentID:    agentID,
	})
	if err != nil {
		return resp, fmt.Errorf("%w: %v", ErrModelFailure, err)
	}
	assistantMsg, err := s.store.CreateMessage(ctx, NewMessage{SessionID: current.Session.ID, Role: RoleAssistant, Content: reply})
	if err != nil {
		return resp, err
	}
	resp.AssistantMessage = &assistantMsg
	return resp, nil
}

func (s *Service) session(ctx context.Context, userID, agentID, sessionID int64, title string) (Session, error) {
	if sessionID > 0 {
		return s.store.GetSession(ctx, userID, agentID, sessionID)
	}
	return s.store.GetLatestOrCreateSession(ctx, userID, agentID, title)
}

func makeSessionTitle(content string) string {
	fields := strings.Fields(strings.TrimSpace(content))
	if len(fields) == 0 {
		return "新对话"
	}
	title := strings.Join(fields, " ")
	runes := []rune(title)
	if len(runes) > 28 {
		title = string(runes[:28]) + "..."
	}
	return title
}

func buildPrompt(agent Agent, messages []Message, latest string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "你正在和用户进行一对一对话。保持%s的人设，直接回答用户。\n", agent.Name)
	for _, msg := range orderMessages(messages) {
		fmt.Fprintf(&b, "%s: %s\n", msg.Role, msg.Content)
	}
	fmt.Fprintf(&b, "%s: %s", RoleUser, latest)
	return b.String()
}

func orderMessages(messages []Message) []Message {
	if messages == nil {
		return []Message{}
	}
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].ID < messages[j].ID })
	return messages
}

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sqlx.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) ListSessions(ctx context.Context, userID int64) ([]SessionSummary, error) {
	var rows []struct {
		Session
		AgentID      int64  `db:"agent_id"`
		AgentName    string `db:"agent_name"`
		SystemPrompt string `db:"system_prompt"`
		LastMessage  string `db:"last_message"`
		MessageCount int64  `db:"message_count"`
	}
	err := s.db.SelectContext(ctx, &rows, `
		SELECT
			s.id,
			s.user_id,
			s.ai_agent_id,
			s.title,
			s.created_at,
			s.updated_at,
			a.id AS agent_id,
			a.name AS agent_name,
			COALESCE(a.system_prompt, '') AS system_prompt,
			COALESCE((
				SELECT m.content
				FROM ai_chat_messages m
				WHERE m.session_id = s.id
				ORDER BY m.id DESC
				LIMIT 1
			), '') AS last_message,
			(
				SELECT COUNT(*)
				FROM ai_chat_messages m
				WHERE m.session_id = s.id
			) AS message_count
		FROM ai_chat_sessions s
		JOIN ai_agents a ON a.id = s.ai_agent_id
		WHERE s.user_id = ?
		ORDER BY s.updated_at DESC, s.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	out := make([]SessionSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, SessionSummary{
			Session:      row.Session,
			Agent:        Agent{ID: row.AgentID, Name: row.AgentName, SystemPrompt: row.SystemPrompt},
			LastMessage:  row.LastMessage,
			MessageCount: row.MessageCount,
		})
	}
	return out, nil
}

func (s *SQLStore) GetAgent(ctx context.Context, id int64) (Agent, error) {
	var agent Agent
	err := s.db.GetContext(ctx, &agent, `
		SELECT id, name, COALESCE(system_prompt, '') AS system_prompt
		FROM ai_agents
		WHERE id = ? AND enabled = TRUE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Agent{}, ErrAgentNotFound
	}
	return agent, err
}

func (s *SQLStore) CreateSession(ctx context.Context, userID, agentID int64, title string) (Session, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO ai_chat_sessions (user_id, ai_agent_id, title)
		VALUES (?, ?, ?)`,
		userID, agentID, title)
	if err != nil {
		return Session{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Session{}, err
	}
	var session Session
	err = s.db.GetContext(ctx, &session, `
		SELECT id, user_id, ai_agent_id, title, created_at, updated_at
		FROM ai_chat_sessions
		WHERE id = ?`, id)
	return session, err
}

func (s *SQLStore) GetLatestOrCreateSession(ctx context.Context, userID, agentID int64, title string) (Session, error) {
	var session Session
	err := s.db.GetContext(ctx, &session, `
		SELECT id, user_id, ai_agent_id, title, created_at, updated_at
		FROM ai_chat_sessions
		WHERE user_id = ? AND ai_agent_id = ?
		ORDER BY updated_at DESC, id DESC
		LIMIT 1`, userID, agentID)
	if err == nil {
		return session, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Session{}, err
	}
	return s.CreateSession(ctx, userID, agentID, title)
}

func (s *SQLStore) GetSession(ctx context.Context, userID, agentID, sessionID int64) (Session, error) {
	var session Session
	err := s.db.GetContext(ctx, &session, `
		SELECT id, user_id, ai_agent_id, title, created_at, updated_at
		FROM ai_chat_sessions
		WHERE id = ? AND user_id = ? AND ai_agent_id = ?`, sessionID, userID, agentID)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrAgentNotFound
	}
	return session, err
}

func (s *SQLStore) UpdateSessionTitle(ctx context.Context, sessionID int64, title string) (Session, error) {
	if _, err := s.db.ExecContext(ctx, `
		UPDATE ai_chat_sessions
		SET title = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, title, sessionID); err != nil {
		return Session{}, err
	}
	var session Session
	err := s.db.GetContext(ctx, &session, `
		SELECT id, user_id, ai_agent_id, title, created_at, updated_at
		FROM ai_chat_sessions
		WHERE id = ?`, sessionID)
	return session, err
}

func (s *SQLStore) ListMessages(ctx context.Context, sessionID int64) ([]Message, error) {
	var messages []Message
	err := s.db.SelectContext(ctx, &messages, `
		SELECT id, session_id, role, content, error_message, created_at
		FROM ai_chat_messages
		WHERE session_id = ?
		ORDER BY id ASC`, sessionID)
	return messages, err
}

func (s *SQLStore) CreateMessage(ctx context.Context, in NewMessage) (Message, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO ai_chat_messages (session_id, role, content)
		VALUES (?, ?, ?)`, in.SessionID, in.Role, in.Content)
	if err != nil {
		return Message{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Message{}, err
	}
	var msg Message
	err = s.db.GetContext(ctx, &msg, `
		SELECT id, session_id, role, content, error_message, created_at
		FROM ai_chat_messages
		WHERE id = ?`, id)
	return msg, err
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	resp, err := h.service.List(r.Context(), sub.UserID)
	writeChatResponse(w, resp, err)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID, ok := parseAgentID(w, r)
	if !ok {
		return
	}
	resp, err := h.service.Create(r.Context(), sub.UserID, agentID)
	writeChatResponse(w, resp, err)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID, ok := parseAgentID(w, r)
	if !ok {
		return
	}
	sessionID, ok := parseOptionalSessionID(w, r)
	if !ok {
		return
	}
	resp, err := h.service.Get(r.Context(), sub.UserID, agentID, sessionID)
	writeChatResponse(w, resp, err)
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID, ok := parseAgentID(w, r)
	if !ok {
		return
	}
	var req struct {
		Content   string `json:"content"`
		SessionID int64  `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	resp, err := h.service.Send(r.Context(), sub.UserID, agentID, req.SessionID, req.Content)
	writeChatResponse(w, resp, err)
}

func parseAgentID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("agentId"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid agent id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func parseOptionalSessionID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.URL.Query().Get("sessionId")
	if raw == "" {
		return 0, true
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func writeChatResponse(w http.ResponseWriter, payload any, err error) {
	if err == nil {
		_ = json.NewEncoder(w).Encode(payload)
		return
	}
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrEmptyMessage):
		status = http.StatusBadRequest
	case errors.Is(err, ErrAgentNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrModelFailure):
		status = http.StatusBadGateway
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error(), "data": payload})
}
