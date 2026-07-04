package chat

import (
	"context"
	"errors"
	"testing"

	"ai-forum/backend/internal/ai/modelclient"
)

func TestServiceGetsOrCreatesSessionWithOrderedMessages(t *testing.T) {
	store := newFakeStore()
	store.messages = []Message{
		{ID: 2, SessionID: 9, Role: RoleAssistant, Content: "第二条"},
		{ID: 1, SessionID: 9, Role: RoleUser, Content: "第一条"},
	}
	svc := NewService(store, &fakeModel{reply: "unused"})

	got, err := svc.Get(context.Background(), 7, 1001)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Session.UserID != 7 || got.Session.AIAgentID != 1001 {
		t.Fatalf("session = %#v, want user 7 agent 1001", got.Session)
	}
	if len(got.Messages) != 2 || got.Messages[0].ID != 1 || got.Messages[1].ID != 2 {
		t.Fatalf("messages not ordered by id: %#v", got.Messages)
	}
}

func TestServiceReturnsEmptyMessagesSlice(t *testing.T) {
	svc := NewService(newFakeStore(), &fakeModel{reply: "unused"})

	got, err := svc.Get(context.Background(), 7, 1001)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Messages == nil || len(got.Messages) != 0 {
		t.Fatalf("messages = %#v, want empty non-nil slice", got.Messages)
	}
}

func TestServiceListsUserSessions(t *testing.T) {
	store := newFakeStore()
	store.sessions = []SessionSummary{
		{Session: Session{ID: 2, UserID: 7, AIAgentID: 1002, Title: "赵务实"}, Agent: Agent{ID: 1002, Name: "赵务实"}, LastMessage: "第二段", MessageCount: 4},
		{Session: Session{ID: 1, UserID: 7, AIAgentID: 1001, Title: "林理臣"}, Agent: Agent{ID: 1001, Name: "林理臣"}, LastMessage: "第一段", MessageCount: 2},
	}
	svc := NewService(store, &fakeModel{reply: "unused"})

	got, err := svc.List(context.Background(), 7)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(got) != 2 || got[0].Session.ID != 2 || got[1].Session.ID != 1 {
		t.Fatalf("sessions = %#v", got)
	}
	if store.listUserID != 7 {
		t.Fatalf("list user id = %d, want 7", store.listUserID)
	}
}

func TestServiceRejectsEmptyMessage(t *testing.T) {
	svc := NewService(newFakeStore(), &fakeModel{reply: "unused"})

	_, err := svc.Send(context.Background(), 7, 1001, "  \n\t ")
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("Send error = %v, want ErrEmptyMessage", err)
	}
}

func TestServicePersistsUserAndAssistantMessages(t *testing.T) {
	store := newFakeStore()
	model := &fakeModel{reply: "你好，我是林理臣。"}
	svc := NewService(store, model)

	got, err := svc.Send(context.Background(), 7, 1001, "聊聊增长")
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if got.UserMessage.Role != RoleUser || got.UserMessage.Content != "聊聊增长" {
		t.Fatalf("user message = %#v", got.UserMessage)
	}
	if got.AssistantMessage == nil || got.AssistantMessage.Role != RoleAssistant || got.AssistantMessage.Content != "你好，我是林理臣。" {
		t.Fatalf("assistant message = %#v", got.AssistantMessage)
	}
	if len(store.created) != 2 || store.created[0].Role != RoleUser || store.created[1].Role != RoleAssistant {
		t.Fatalf("created messages = %#v, want user then assistant", store.created)
	}
	if model.last.AIAgentID != 1001 || model.last.TaskType != "ai_chat" {
		t.Fatalf("model request metadata = %#v", model.last)
	}
}

func TestServiceKeepsUserMessageWhenModelFails(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, &fakeModel{err: errors.New("model unavailable")})

	got, err := svc.Send(context.Background(), 7, 1001, "还在吗")
	if err == nil {
		t.Fatal("Send returned nil error, want model failure")
	}
	if got.UserMessage.ID == 0 || got.UserMessage.Role != RoleUser {
		t.Fatalf("user message not preserved: %#v", got.UserMessage)
	}
	if got.AssistantMessage != nil {
		t.Fatalf("assistant message = %#v, want nil", got.AssistantMessage)
	}
	if len(store.created) != 1 || store.created[0].Role != RoleUser {
		t.Fatalf("created messages = %#v, want only user message", store.created)
	}
}

type fakeModel struct {
	reply string
	err   error
	last  modelclient.Request
}

func (m *fakeModel) Generate(_ context.Context, in modelclient.Request) (string, error) {
	m.last = in
	if m.err != nil {
		return "", m.err
	}
	return m.reply, nil
}

type fakeStore struct {
	session    Session
	sessions   []SessionSummary
	agent      Agent
	messages   []Message
	created    []Message
	listUserID int64
	nextID     int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		session: Session{ID: 9, UserID: 7, AIAgentID: 1001, Title: "林理臣"},
		agent:   Agent{ID: 1001, Name: "林理臣", SystemPrompt: "你是林理臣。"},
		nextID:  100,
	}
}

func (s *fakeStore) GetAgent(_ context.Context, id int64) (Agent, error) {
	if id != s.agent.ID {
		return Agent{}, ErrAgentNotFound
	}
	return s.agent, nil
}

func (s *fakeStore) ListSessions(_ context.Context, userID int64) ([]SessionSummary, error) {
	s.listUserID = userID
	return s.sessions, nil
}

func (s *fakeStore) GetOrCreateSession(_ context.Context, userID, agentID int64, title string) (Session, error) {
	s.session.UserID = userID
	s.session.AIAgentID = agentID
	if s.session.Title == "" {
		s.session.Title = title
	}
	return s.session, nil
}

func (s *fakeStore) ListMessages(_ context.Context, sessionID int64) ([]Message, error) {
	out := append([]Message(nil), s.messages...)
	return orderMessages(out), nil
}

func (s *fakeStore) CreateMessage(_ context.Context, in NewMessage) (Message, error) {
	s.nextID++
	msg := Message{ID: s.nextID, SessionID: in.SessionID, Role: in.Role, Content: in.Content}
	s.created = append(s.created, msg)
	s.messages = append(s.messages, msg)
	return msg, nil
}
