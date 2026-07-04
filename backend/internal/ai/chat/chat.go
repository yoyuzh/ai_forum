// Package chat owns one-to-one user conversations with AI agents.
package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/ai/modelclient"
	"ai-forum/backend/internal/auth"
)

const (
	RoleUser      = "USER"
	RoleAssistant = "AI"

	StatusActive  = "ACTIVE"
	StatusDeleted = "DELETED"

	MessageDone      = "DONE"
	MessageStreaming = "STREAMING"
	MessageFailed    = "FAILED"
	MessagePartial   = "PARTIAL"
)

var (
	ErrAgentNotFound   = errors.New("agent not found")
	ErrSessionNotFound = errors.New("session not found")
	ErrMessageNotFound = errors.New("message not found")
	ErrEmptyMessage    = errors.New("empty message")
	ErrRequestID       = errors.New("request id required")
	ErrModelFailure    = errors.New("model failure")
)

type Agent struct {
	ID           int64  `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	SystemPrompt string `json:"systemPrompt" db:"system_prompt"`
}

type Session struct {
	ID                 int64  `json:"id" db:"id"`
	UserID             int64  `json:"userId" db:"user_id"`
	AIAgentID          int64  `json:"aiAgentId" db:"ai_agent_id"`
	Title              string `json:"title" db:"title"`
	Status             string `json:"status" db:"status"`
	LastMessagePreview string `json:"lastMessagePreview" db:"last_message_preview"`
	MessageCount       int64  `json:"messageCount" db:"message_count"`
	CreatedAt          string `json:"createdAt" db:"created_at"`
	UpdatedAt          string `json:"updatedAt" db:"updated_at"`
}

type SessionSummary struct {
	Session      Session `json:"session"`
	Agent        Agent   `json:"agent"`
	LastMessage  string  `json:"lastMessage"`
	MessageCount int64   `json:"messageCount"`
}

type PagedSessions struct {
	Items    []SessionSummary `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Total    int64            `json:"total"`
}

type Message struct {
	ID           int64   `json:"id" db:"id"`
	SessionID    int64   `json:"sessionId" db:"session_id"`
	Role         string  `json:"role" db:"role"`
	SenderType   string  `json:"senderType" db:"sender_type"`
	Content      string  `json:"content" db:"content"`
	Status       string  `json:"status" db:"status"`
	SequenceNo   int64   `json:"sequenceNo" db:"sequence_no"`
	RequestID    *string `json:"requestId,omitempty" db:"request_id"`
	ErrorMessage *string `json:"errorMessage,omitempty" db:"error_message"`
	CreatedAt    string  `json:"createdAt" db:"created_at"`
	UpdatedAt    string  `json:"updatedAt" db:"updated_at"`
}

type ListOptions struct {
	Page     int
	PageSize int
	AgentID  int64
}

type SendInput struct {
	SessionID int64
	AgentID   int64
	Content   string
	RequestID string
}

type RetryInput struct {
	MessageID int64
	RequestID string
}

type SendStart struct {
	Session          Session
	Agent            Agent
	UserMessage      Message
	AssistantMessage Message
	CreatedSession   bool
	DuplicateRequest bool
	History          []Message
}

type RetryStart struct {
	Session          Session
	Agent            Agent
	AssistantMessage Message
	DuplicateRequest bool
	History          []Message
}

type Store interface {
	ListSessions(context.Context, int64, ListOptions) (PagedSessions, error)
	GetAgent(context.Context, int64) (Agent, error)
	GetSession(context.Context, int64, int64) (Session, error)
	ListMessages(context.Context, int64, int64) ([]Message, error)
	StartSend(context.Context, int64, SendInput) (SendStart, error)
	CompleteAssistant(context.Context, int64, int64, string) (Message, Session, error)
	FailAssistant(context.Context, int64, int64, string, string) (Message, error)
	StartRetry(context.Context, int64, RetryInput) (RetryStart, error)
	DeleteSession(context.Context, int64, int64) error
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

type StreamEvent struct {
	Event string
	Data  any
}

type SendResponse struct {
	Session          Session `json:"session"`
	UserMessage      Message `json:"userMessage"`
	AssistantMessage Message `json:"assistantMessage"`
}

func (s *Service) List(ctx context.Context, userID int64, opts ListOptions) (PagedSessions, error) {
	normalizeListOptions(&opts)
	return s.store.ListSessions(ctx, userID, opts)
}

func (s *Service) Get(ctx context.Context, userID, sessionID int64) (SessionResponse, error) {
	session, err := s.store.GetSession(ctx, userID, sessionID)
	if err != nil {
		return SessionResponse{}, err
	}
	agent, err := s.store.GetAgent(ctx, session.AIAgentID)
	if err != nil {
		return SessionResponse{}, err
	}
	messages, err := s.store.ListMessages(ctx, userID, session.ID)
	if err != nil {
		return SessionResponse{}, err
	}
	return SessionResponse{Session: session, Agent: agent, Messages: normalizeMessages(messages)}, nil
}

func (s *Service) Send(ctx context.Context, userID int64, in SendInput, emit func(StreamEvent) error) (SendResponse, error) {
	in.Content = strings.TrimSpace(in.Content)
	in.RequestID = strings.TrimSpace(in.RequestID)
	if in.Content == "" {
		return SendResponse{}, ErrEmptyMessage
	}
	if in.AgentID <= 0 {
		return SendResponse{}, ErrAgentNotFound
	}
	if in.RequestID == "" {
		return SendResponse{}, ErrRequestID
	}

	start, err := s.store.StartSend(ctx, userID, in)
	if err != nil {
		return SendResponse{}, err
	}
	if start.CreatedSession {
		if err := emit(StreamEvent{Event: "conversation_created", Data: map[string]any{"conversationId": start.Session.ID, "session": start.Session}}); err != nil {
			return SendResponse{}, err
		}
	}
	if err := emit(StreamEvent{Event: "user_message_saved", Data: map[string]any{"message": start.UserMessage, "messageId": start.UserMessage.ID, "sequenceNo": start.UserMessage.SequenceNo}}); err != nil {
		return SendResponse{}, err
	}
	if err := emit(StreamEvent{Event: "ai_message_created", Data: map[string]any{"message": start.AssistantMessage, "messageId": start.AssistantMessage.ID, "sequenceNo": start.AssistantMessage.SequenceNo}}); err != nil {
		return SendResponse{}, err
	}
	if start.DuplicateRequest && start.AssistantMessage.Status != MessageStreaming {
		_ = emit(StreamEvent{Event: "done", Data: map[string]any{"message": start.AssistantMessage, "messageId": start.AssistantMessage.ID, "status": start.AssistantMessage.Status, "session": start.Session}})
		return SendResponse{Session: start.Session, UserMessage: start.UserMessage, AssistantMessage: start.AssistantMessage}, nil
	}

	reply, err := s.model.Generate(ctx, modelclient.Request{
		SystemPrompt: start.Agent.SystemPrompt,
		Prompt:       buildPrompt(start.Agent, start.History, in.Content),
		TaskType:     "ai_chat",
		AIAgentID:    start.Agent.ID,
	})
	if err != nil {
		failed, _ := s.store.FailAssistant(ctx, userID, start.AssistantMessage.ID, MessageFailed, err.Error())
		_ = emit(StreamEvent{Event: "error", Data: map[string]any{"code": "MODEL_FAILURE", "message": "AI 回复生成失败", "messageId": start.AssistantMessage.ID, "aiMessage": failed}})
		return SendResponse{Session: start.Session, UserMessage: start.UserMessage, AssistantMessage: failed}, fmt.Errorf("%w: %v", ErrModelFailure, err)
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		failed, _ := s.store.FailAssistant(ctx, userID, start.AssistantMessage.ID, MessageFailed, "empty model response")
		_ = emit(StreamEvent{Event: "error", Data: map[string]any{"code": "EMPTY_MODEL_RESPONSE", "message": "AI 回复为空", "messageId": start.AssistantMessage.ID, "aiMessage": failed}})
		return SendResponse{Session: start.Session, UserMessage: start.UserMessage, AssistantMessage: failed}, ErrModelFailure
	}
	for _, chunk := range chunks(reply, 16) {
		if err := emit(StreamEvent{Event: "token", Data: map[string]string{"content": chunk}}); err != nil {
			partial, _ := s.store.FailAssistant(ctx, userID, start.AssistantMessage.ID, MessagePartial, "client disconnected")
			return SendResponse{Session: start.Session, UserMessage: start.UserMessage, AssistantMessage: partial}, err
		}
	}
	assistant, session, err := s.store.CompleteAssistant(ctx, userID, start.AssistantMessage.ID, reply)
	if err != nil {
		return SendResponse{}, err
	}
	if err := emit(StreamEvent{Event: "done", Data: map[string]any{"message": assistant, "messageId": assistant.ID, "status": assistant.Status, "session": session}}); err != nil {
		return SendResponse{}, err
	}
	return SendResponse{Session: session, UserMessage: start.UserMessage, AssistantMessage: assistant}, nil
}

func (s *Service) Retry(ctx context.Context, userID int64, in RetryInput, emit func(StreamEvent) error) (Message, error) {
	in.RequestID = strings.TrimSpace(in.RequestID)
	if in.MessageID <= 0 {
		return Message{}, ErrMessageNotFound
	}
	if in.RequestID == "" {
		return Message{}, ErrRequestID
	}
	start, err := s.store.StartRetry(ctx, userID, in)
	if err != nil {
		return Message{}, err
	}
	if err := emit(StreamEvent{Event: "ai_message_created", Data: map[string]any{"message": start.AssistantMessage, "messageId": start.AssistantMessage.ID, "sequenceNo": start.AssistantMessage.SequenceNo}}); err != nil {
		return Message{}, err
	}
	if start.DuplicateRequest && start.AssistantMessage.Status != MessageStreaming {
		_ = emit(StreamEvent{Event: "done", Data: map[string]any{"message": start.AssistantMessage, "messageId": start.AssistantMessage.ID, "status": start.AssistantMessage.Status, "session": start.Session}})
		return start.AssistantMessage, nil
	}
	latest := latestUserContent(start.History)
	reply, err := s.model.Generate(ctx, modelclient.Request{
		SystemPrompt: start.Agent.SystemPrompt,
		Prompt:       buildPrompt(start.Agent, withoutMessage(start.History, start.AssistantMessage.ID), latest),
		TaskType:     "ai_chat_retry",
		AIAgentID:    start.Agent.ID,
	})
	if err != nil {
		failed, _ := s.store.FailAssistant(ctx, userID, start.AssistantMessage.ID, MessageFailed, err.Error())
		_ = emit(StreamEvent{Event: "error", Data: map[string]any{"code": "MODEL_FAILURE", "message": "AI 回复生成失败", "messageId": start.AssistantMessage.ID, "aiMessage": failed}})
		return failed, fmt.Errorf("%w: %v", ErrModelFailure, err)
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		failed, _ := s.store.FailAssistant(ctx, userID, start.AssistantMessage.ID, MessageFailed, "empty model response")
		_ = emit(StreamEvent{Event: "error", Data: map[string]any{"code": "EMPTY_MODEL_RESPONSE", "message": "AI 回复为空", "messageId": start.AssistantMessage.ID, "aiMessage": failed}})
		return failed, ErrModelFailure
	}
	for _, chunk := range chunks(reply, 16) {
		if err := emit(StreamEvent{Event: "token", Data: map[string]string{"content": chunk}}); err != nil {
			partial, _ := s.store.FailAssistant(ctx, userID, start.AssistantMessage.ID, MessagePartial, "client disconnected")
			return partial, err
		}
	}
	assistant, session, err := s.store.CompleteAssistant(ctx, userID, start.AssistantMessage.ID, reply)
	if err != nil {
		return Message{}, err
	}
	if err := emit(StreamEvent{Event: "done", Data: map[string]any{"message": assistant, "messageId": assistant.ID, "status": assistant.Status, "session": session}}); err != nil {
		return Message{}, err
	}
	return assistant, nil
}

func (s *Service) Delete(ctx context.Context, userID, sessionID int64) error {
	return s.store.DeleteSession(ctx, userID, sessionID)
}

func buildPrompt(agent Agent, messages []Message, latest string) string {
	var b strings.Builder
	if strings.TrimSpace(agent.SystemPrompt) != "" {
		b.WriteString(agent.SystemPrompt)
	} else {
		fmt.Fprintf(&b, "你是%s。", agent.Name)
	}
	b.WriteString("\n\n你正在和用户进行一对一对话。保持该角色的人设、说话风格和禁止事项。不要说自己是程序。")
	for _, msg := range normalizeMessages(messages) {
		if msg.Status == MessageFailed || msg.Status == MessagePartial || msg.ID == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n%s: %s", msg.Role, msg.Content)
	}
	if strings.TrimSpace(latest) != "" {
		fmt.Fprintf(&b, "\n%s: %s", RoleUser, latest)
	}
	return b.String()
}

func normalizeMessages(messages []Message) []Message {
	if messages == nil {
		return []Message{}
	}
	for i := range messages {
		messages[i].SenderType = messages[i].Role
	}
	return messages
}

func normalizeListOptions(opts *ListOptions) {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 20
	}
	if opts.PageSize > 100 {
		opts.PageSize = 100
	}
}

func makeSessionTitle(content string) string {
	title := strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
	if title == "" {
		return "新对话"
	}
	return firstRunes(title, 20)
}

func preview(content string) string {
	return firstRunes(strings.Join(strings.Fields(strings.TrimSpace(content)), " "), 255)
}

func firstRunes(value string, max int) string {
	if utf8.RuneCountInString(value) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max])
}

func chunks(value string, size int) []string {
	runes := []rune(value)
	out := make([]string, 0, (len(runes)+size-1)/size)
	for len(runes) > 0 {
		n := size
		if len(runes) < n {
			n = len(runes)
		}
		out = append(out, string(runes[:n]))
		runes = runes[n:]
	}
	return out
}

func latestUserContent(messages []Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == RoleUser {
			return messages[i].Content
		}
	}
	return ""
}

func withoutMessage(messages []Message, id int64) []Message {
	out := make([]Message, 0, len(messages))
	for _, msg := range messages {
		if msg.ID != id {
			out = append(out, msg)
		}
	}
	return out
}

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sqlx.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) ListSessions(ctx context.Context, userID int64, opts ListOptions) (PagedSessions, error) {
	normalizeListOptions(&opts)
	args := []any{userID}
	agentClause := ""
	if opts.AgentID > 0 {
		agentClause = " AND s.ai_agent_id = ?"
		args = append(args, opts.AgentID)
	}
	var total int64
	if err := s.db.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM ai_chat_sessions s
		WHERE s.user_id = ? AND s.status = 'ACTIVE' AND s.message_count > 0`+agentClause, args...); err != nil {
		return PagedSessions{}, err
	}
	args = append(args, opts.PageSize, (opts.Page-1)*opts.PageSize)
	var rows []struct {
		Session
		AgentID      int64  `db:"agent_id"`
		AgentName    string `db:"agent_name"`
		SystemPrompt string `db:"system_prompt"`
	}
	err := s.db.SelectContext(ctx, &rows, `
		SELECT
			s.id, s.user_id, s.ai_agent_id, s.title, s.status, COALESCE(s.last_message_preview, '') AS last_message_preview,
			s.message_count, s.created_at, s.updated_at,
			a.id AS agent_id, a.name AS agent_name, COALESCE(a.system_prompt, '') AS system_prompt
		FROM ai_chat_sessions s
		JOIN ai_agents a ON a.id = s.ai_agent_id
		WHERE s.user_id = ? AND s.status = 'ACTIVE' AND s.message_count > 0`+agentClause+`
		ORDER BY s.updated_at DESC, s.id DESC
		LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return PagedSessions{}, err
	}
	out := make([]SessionSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, SessionSummary{
			Session:      row.Session,
			Agent:        Agent{ID: row.AgentID, Name: row.AgentName, SystemPrompt: row.SystemPrompt},
			LastMessage:  row.LastMessagePreview,
			MessageCount: row.MessageCount,
		})
	}
	return PagedSessions{Items: out, Page: opts.Page, PageSize: opts.PageSize, Total: total}, nil
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

func (s *SQLStore) GetSession(ctx context.Context, userID, sessionID int64) (Session, error) {
	var session Session
	err := s.db.GetContext(ctx, &session, sessionSelect()+` WHERE id = ? AND user_id = ? AND status = 'ACTIVE'`, sessionID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	return session, err
}

func (s *SQLStore) ListMessages(ctx context.Context, userID, sessionID int64) ([]Message, error) {
	var messages []Message
	err := s.db.SelectContext(ctx, &messages, messageSelect()+`
		WHERE m.session_id = ? AND s.user_id = ? AND s.status = 'ACTIVE'
		ORDER BY m.sequence_no ASC`, sessionID, userID)
	return normalizeMessages(messages), err
}

func (s *SQLStore) StartSend(ctx context.Context, userID int64, in SendInput) (SendStart, error) {
	if existing, ok, err := s.sendByRequestID(ctx, userID, in.RequestID); ok || err != nil {
		return existing, err
	}

	agent, err := s.GetAgent(ctx, in.AgentID)
	if err != nil {
		return SendStart{}, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return SendStart{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var session Session
	created := false
	if in.SessionID > 0 {
		err = tx.GetContext(ctx, &session, sessionSelect()+` WHERE id = ? AND user_id = ? AND status = 'ACTIVE' FOR UPDATE`, in.SessionID, userID)
		if errors.Is(err, sql.ErrNoRows) {
			return SendStart{}, ErrSessionNotFound
		}
		if err != nil {
			return SendStart{}, err
		}
		if session.AIAgentID != in.AgentID {
			return SendStart{}, ErrSessionNotFound
		}
	} else {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO ai_chat_sessions (user_id, ai_agent_id, title, status, last_message_preview, message_count)
			VALUES (?, ?, ?, 'ACTIVE', ?, 0)`, userID, in.AgentID, makeSessionTitle(in.Content), preview(in.Content))
		if err != nil {
			return SendStart{}, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return SendStart{}, err
		}
		err = tx.GetContext(ctx, &session, sessionSelect()+` WHERE id = ? FOR UPDATE`, id)
		if err != nil {
			return SendStart{}, err
		}
		created = true
	}

	nextSeq, err := nextSequence(ctx, tx, session.ID)
	if err != nil {
		return SendStart{}, err
	}
	userMsg, err := insertMessage(ctx, tx, session.ID, RoleUser, in.Content, MessageDone, nextSeq, &in.RequestID)
	if err != nil {
		if existing, ok, fetchErr := s.sendByRequestID(ctx, userID, in.RequestID); ok || fetchErr != nil {
			return existing, fetchErr
		}
		return SendStart{}, err
	}
	assistantMsg, err := insertMessage(ctx, tx, session.ID, RoleAssistant, "", MessageStreaming, nextSeq+1, nil)
	if err != nil {
		return SendStart{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE ai_chat_sessions
		SET last_message_preview = ?, message_count = message_count + 2, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, preview(in.Content), session.ID); err != nil {
		return SendStart{}, err
	}
	if err := tx.Commit(); err != nil {
		return SendStart{}, err
	}
	session, err = s.GetSession(ctx, userID, session.ID)
	if err != nil {
		return SendStart{}, err
	}
	history, err := s.ListMessages(ctx, userID, session.ID)
	if err != nil {
		return SendStart{}, err
	}
	return SendStart{
		Session:          session,
		Agent:            agent,
		UserMessage:      userMsg,
		AssistantMessage: assistantMsg,
		CreatedSession:   created,
		History:          withoutMessage(history, assistantMsg.ID),
	}, nil
}

func (s *SQLStore) CompleteAssistant(ctx context.Context, userID, messageID int64, content string) (Message, Session, error) {
	var msg Message
	err := s.db.GetContext(ctx, &msg, messageSelect()+`
		WHERE m.id = ? AND s.user_id = ? AND s.status = 'ACTIVE' AND m.role = 'AI'`, messageID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Message{}, Session{}, ErrMessageNotFound
	}
	if err != nil {
		return Message{}, Session{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE ai_chat_messages
		SET content = ?, status = 'DONE', error_message = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, content, messageID); err != nil {
		return Message{}, Session{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE ai_chat_sessions
		SET last_message_preview = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?`, preview(content), msg.SessionID, userID); err != nil {
		return Message{}, Session{}, err
	}
	var updated Message
	if err := s.db.GetContext(ctx, &updated, messageSelect()+` WHERE m.id = ? AND s.user_id = ?`, messageID, userID); err != nil {
		return Message{}, Session{}, err
	}
	session, err := s.GetSession(ctx, userID, updated.SessionID)
	return normalizeMessages([]Message{updated})[0], session, err
}

func (s *SQLStore) FailAssistant(ctx context.Context, userID, messageID int64, status, errorMessage string) (Message, error) {
	if status != MessagePartial {
		status = MessageFailed
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE ai_chat_messages m
		JOIN ai_chat_sessions s ON s.id = m.session_id
		SET m.status = ?, m.error_message = ?, m.updated_at = CURRENT_TIMESTAMP
		WHERE m.id = ? AND s.user_id = ? AND s.status = 'ACTIVE' AND m.role = 'AI'`,
		status, firstRunes(errorMessage, 500), messageID, userID); err != nil {
		return Message{}, err
	}
	var msg Message
	err := s.db.GetContext(ctx, &msg, messageSelect()+` WHERE m.id = ? AND s.user_id = ?`, messageID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Message{}, ErrMessageNotFound
	}
	if err != nil {
		return Message{}, err
	}
	return normalizeMessages([]Message{msg})[0], nil
}

func (s *SQLStore) StartRetry(ctx context.Context, userID int64, in RetryInput) (RetryStart, error) {
	if existing, ok, err := s.retryByRequestID(ctx, userID, in.RequestID); ok || err != nil {
		return existing, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return RetryStart{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var msg Message
	err = tx.GetContext(ctx, &msg, messageSelect()+`
		WHERE m.id = ? AND s.user_id = ? AND s.status = 'ACTIVE' AND m.role = 'AI' AND m.status IN ('FAILED', 'PARTIAL')
		FOR UPDATE`, in.MessageID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return RetryStart{}, ErrMessageNotFound
	}
	if err != nil {
		return RetryStart{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE ai_chat_messages
		SET status = 'STREAMING', content = '', error_message = NULL, request_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, in.RequestID, msg.ID); err != nil {
		if existing, ok, fetchErr := s.retryByRequestID(ctx, userID, in.RequestID); ok || fetchErr != nil {
			return existing, fetchErr
		}
		return RetryStart{}, err
	}
	if err := tx.Commit(); err != nil {
		return RetryStart{}, err
	}
	session, err := s.GetSession(ctx, userID, msg.SessionID)
	if err != nil {
		return RetryStart{}, err
	}
	agent, err := s.GetAgent(ctx, session.AIAgentID)
	if err != nil {
		return RetryStart{}, err
	}
	history, err := s.ListMessages(ctx, userID, session.ID)
	if err != nil {
		return RetryStart{}, err
	}
	msg.Status = MessageStreaming
	msg.Content = ""
	msg.ErrorMessage = nil
	msg.RequestID = &in.RequestID
	return RetryStart{Session: session, Agent: agent, AssistantMessage: normalizeMessages([]Message{msg})[0], History: history}, nil
}

func (s *SQLStore) DeleteSession(ctx context.Context, userID, sessionID int64) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE ai_chat_sessions
		SET status = 'DELETED', updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND status = 'ACTIVE'`, sessionID, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (s *SQLStore) sendByRequestID(ctx context.Context, userID int64, requestID string) (SendStart, bool, error) {
	if requestID == "" {
		return SendStart{}, false, nil
	}
	var userMsg Message
	err := s.db.GetContext(ctx, &userMsg, messageSelect()+`
		WHERE m.request_id = ? AND s.user_id = ? AND s.status = 'ACTIVE'`, requestID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return SendStart{}, false, nil
	}
	if err != nil {
		return SendStart{}, false, err
	}
	session, err := s.GetSession(ctx, userID, userMsg.SessionID)
	if err != nil {
		return SendStart{}, false, err
	}
	agent, err := s.GetAgent(ctx, session.AIAgentID)
	if err != nil {
		return SendStart{}, false, err
	}
	var aiMsg Message
	err = s.db.GetContext(ctx, &aiMsg, messageSelect()+`
		WHERE m.session_id = ? AND m.sequence_no = ? AND s.user_id = ?`, userMsg.SessionID, userMsg.SequenceNo+1, userID)
	if errors.Is(err, sql.ErrNoRows) {
		aiMsg = Message{SessionID: userMsg.SessionID, Role: RoleAssistant, SenderType: RoleAssistant, Status: MessageFailed, SequenceNo: userMsg.SequenceNo + 1}
	} else if err != nil {
		return SendStart{}, false, err
	}
	history, err := s.ListMessages(ctx, userID, session.ID)
	if err != nil {
		return SendStart{}, false, err
	}
	return SendStart{Session: session, Agent: agent, UserMessage: normalizeMessages([]Message{userMsg})[0], AssistantMessage: normalizeMessages([]Message{aiMsg})[0], DuplicateRequest: true, History: history}, true, nil
}

func (s *SQLStore) retryByRequestID(ctx context.Context, userID int64, requestID string) (RetryStart, bool, error) {
	if requestID == "" {
		return RetryStart{}, false, nil
	}
	var msg Message
	err := s.db.GetContext(ctx, &msg, messageSelect()+`
		WHERE m.request_id = ? AND s.user_id = ? AND s.status = 'ACTIVE' AND m.role = 'AI'`, requestID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return RetryStart{}, false, nil
	}
	if err != nil {
		return RetryStart{}, false, err
	}
	session, err := s.GetSession(ctx, userID, msg.SessionID)
	if err != nil {
		return RetryStart{}, false, err
	}
	agent, err := s.GetAgent(ctx, session.AIAgentID)
	if err != nil {
		return RetryStart{}, false, err
	}
	history, err := s.ListMessages(ctx, userID, session.ID)
	if err != nil {
		return RetryStart{}, false, err
	}
	return RetryStart{Session: session, Agent: agent, AssistantMessage: normalizeMessages([]Message{msg})[0], DuplicateRequest: true, History: history}, true, nil
}

func sessionSelect() string {
	return `SELECT id, user_id, ai_agent_id, title, COALESCE(status, 'ACTIVE') AS status,
		COALESCE(last_message_preview, '') AS last_message_preview, COALESCE(message_count, 0) AS message_count,
		created_at, updated_at FROM ai_chat_sessions`
}

func messageSelect() string {
	return `SELECT m.id, m.session_id, m.role, m.role AS sender_type, m.content, COALESCE(m.status, 'DONE') AS status,
		COALESCE(m.sequence_no, m.id) AS sequence_no, m.request_id, m.error_message, m.created_at,
		COALESCE(m.updated_at, m.created_at) AS updated_at
		FROM ai_chat_messages m JOIN ai_chat_sessions s ON s.id = m.session_id`
}

func nextSequence(ctx context.Context, tx *sqlx.Tx, sessionID int64) (int64, error) {
	var next int64
	err := tx.GetContext(ctx, &next, `SELECT COALESCE(MAX(sequence_no), 0) + 1 FROM ai_chat_messages WHERE session_id = ?`, sessionID)
	return next, err
}

func insertMessage(ctx context.Context, tx *sqlx.Tx, sessionID int64, role, content, status string, sequenceNo int64, requestID *string) (Message, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO ai_chat_messages (session_id, role, content, status, sequence_no, request_id)
		VALUES (?, ?, ?, ?, ?, ?)`, sessionID, role, content, status, sequenceNo, requestID)
	if err != nil {
		return Message{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Message{}, err
	}
	var msg Message
	if err := tx.GetContext(ctx, &msg, messageSelect()+` WHERE m.id = ?`, id); err != nil {
		return Message{}, err
	}
	return normalizeMessages([]Message{msg})[0], nil
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
	opts := ListOptions{
		Page:     parseIntDefault(r.URL.Query().Get("page"), 1),
		PageSize: parseIntDefault(r.URL.Query().Get("pageSize"), 20),
		AgentID:  parseInt64Default(r.URL.Query().Get("agentId"), 0),
	}
	resp, err := h.service.List(r.Context(), sub.UserID, opts)
	writeChatResponse(w, resp, err)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sessionID, ok := parsePathID(w, r, "conversationId", "invalid conversation id")
	if !ok {
		return
	}
	resp, err := h.service.Get(r.Context(), sub.UserID, sessionID)
	writeChatResponse(w, resp, err)
}

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		ConversationID *int64 `json:"conversationId"`
		AgentID        int64  `json:"agentId"`
		Content        string `json:"content"`
		RequestID      string `json:"requestId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	sessionID := int64(0)
	if req.ConversationID != nil {
		sessionID = *req.ConversationID
	}
	h.writeStream(w, func(emit func(StreamEvent) error) error {
		_, err := h.service.Send(r.Context(), sub.UserID, SendInput{
			SessionID: sessionID,
			AgentID:   req.AgentID,
			Content:   req.Content,
			RequestID: req.RequestID,
		}, emit)
		return err
	})
}

func (h *Handler) Retry(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	messageID, ok := parsePathID(w, r, "messageId", "invalid message id")
	if !ok {
		return
	}
	var req struct {
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	h.writeStream(w, func(emit func(StreamEvent) error) error {
		_, err := h.service.Retry(r.Context(), sub.UserID, RetryInput{MessageID: messageID, RequestID: req.RequestID}, emit)
		return err
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sessionID, ok := parsePathID(w, r, "conversationId", "invalid conversation id")
	if !ok {
		return
	}
	err := h.service.Delete(r.Context(), sub.UserID, sessionID)
	if err != nil {
		writeChatResponse(w, nil, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) writeStream(w http.ResponseWriter, run func(func(StreamEvent) error) error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	emit := func(evt StreamEvent) error {
		payload, err := json.Marshal(evt.Data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Event, payload); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}
	if err := run(emit); err != nil {
		code, message := chatError(err)
		if code >= 500 || code == http.StatusBadRequest || code == http.StatusNotFound {
			_ = emit(StreamEvent{Event: "error", Data: map[string]any{"code": http.StatusText(code), "message": message}})
		}
	}
}

func parsePathID(w http.ResponseWriter, r *http.Request, key, message string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, message, http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func parseIntDefault(raw string, fallback int) int {
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func parseInt64Default(raw string, fallback int64) int64 {
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return v
}

func writeChatResponse(w http.ResponseWriter, payload any, err error) {
	if err == nil {
		_ = json.NewEncoder(w).Encode(payload)
		return
	}
	code, message := chatError(err)
	http.Error(w, message, code)
}

func chatError(err error) (int, string) {
	switch {
	case errors.Is(err, ErrEmptyMessage):
		return http.StatusBadRequest, "empty message"
	case errors.Is(err, ErrRequestID):
		return http.StatusBadRequest, "request id required"
	case errors.Is(err, ErrAgentNotFound):
		return http.StatusBadRequest, "agent not found"
	case errors.Is(err, ErrSessionNotFound), errors.Is(err, ErrMessageNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, ErrModelFailure):
		return http.StatusBadGateway, "model failure"
	default:
		return http.StatusInternalServerError, "chat error"
	}
}

func nowString() string {
	return time.Now().UTC().Format(time.RFC3339)
}
