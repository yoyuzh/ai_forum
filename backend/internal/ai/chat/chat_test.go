package chat

import (
	"context"
	"errors"
	"testing"

	"ai-forum/backend/internal/ai/modelclient"
)

func TestServiceSendsFirstMessageAndStreamsEvents(t *testing.T) {
	store := newFakeStore()
	model := &fakeModel{reply: "你好，我是林理臣。"}
	svc := NewService(store, model)
	var events []StreamEvent

	got, err := svc.Send(context.Background(), 7, SendInput{
		AgentID:   1001,
		Content:   "聊聊增长",
		RequestID: "req-1",
	}, func(evt StreamEvent) error {
		events = append(events, evt)
		return nil
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if got.Session.ID == 0 || got.Session.Title != "聊聊增长" {
		t.Fatalf("session = %#v", got.Session)
	}
	if got.UserMessage.Role != RoleUser || got.UserMessage.SequenceNo != 1 {
		t.Fatalf("user message = %#v", got.UserMessage)
	}
	if got.AssistantMessage.Role != RoleAssistant || got.AssistantMessage.Status != MessageDone || got.AssistantMessage.SequenceNo != 2 {
		t.Fatalf("assistant message = %#v", got.AssistantMessage)
	}
	if len(events) < 5 || events[0].Event != "conversation_created" || events[len(events)-1].Event != "done" {
		t.Fatalf("events = %#v", events)
	}
	if model.last.AIAgentID != 1001 || model.last.TaskType != "ai_chat" {
		t.Fatalf("model request = %#v", model.last)
	}
}

func TestServiceRejectsInvalidSend(t *testing.T) {
	svc := NewService(newFakeStore(), &fakeModel{reply: "unused"})
	_, err := svc.Send(context.Background(), 7, SendInput{AgentID: 1001, Content: "hello"}, noopEmit)
	if !errors.Is(err, ErrRequestID) {
		t.Fatalf("missing request id error = %v, want ErrRequestID", err)
	}
	_, err = svc.Send(context.Background(), 7, SendInput{AgentID: 1001, RequestID: "req", Content: " \n "}, noopEmit)
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("empty message error = %v, want ErrEmptyMessage", err)
	}
}

func TestServiceMarksAssistantFailedWhenModelFails(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, &fakeModel{err: errors.New("model down")})

	got, err := svc.Send(context.Background(), 7, SendInput{
		AgentID:   1001,
		Content:   "还在吗",
		RequestID: "req-2",
	}, noopEmit)
	if err == nil {
		t.Fatal("Send returned nil error, want model failure")
	}
	if got.UserMessage.Role != RoleUser || got.UserMessage.Status != MessageDone {
		t.Fatalf("user message not preserved: %#v", got.UserMessage)
	}
	if got.AssistantMessage.Status != MessageFailed {
		t.Fatalf("assistant status = %q, want FAILED", got.AssistantMessage.Status)
	}
}

func TestServiceRetryReusesAssistantMessage(t *testing.T) {
	store := newFakeStore()
	store.session = Session{ID: 9, UserID: 7, AIAgentID: 1001, Title: "旧会话", Status: StatusActive}
	store.messages = []Message{
		{ID: 1, SessionID: 9, Role: RoleUser, SenderType: RoleUser, Content: "第一问", Status: MessageDone, SequenceNo: 1},
		{ID: 2, SessionID: 9, Role: RoleAssistant, SenderType: RoleAssistant, Content: "坏回复", Status: MessageFailed, SequenceNo: 2},
	}
	svc := NewService(store, &fakeModel{reply: "重试后的回复"})

	got, err := svc.Retry(context.Background(), 7, RetryInput{MessageID: 2, RequestID: "retry-1"}, noopEmit)
	if err != nil {
		t.Fatalf("Retry returned error: %v", err)
	}
	if got.ID != 2 || got.Content != "重试后的回复" || got.Status != MessageDone {
		t.Fatalf("retried message = %#v", got)
	}
	if len(store.messages) != 2 {
		t.Fatalf("message count = %d, want no duplicate user message", len(store.messages))
	}
}

func TestServiceListsPagedSessions(t *testing.T) {
	store := newFakeStore()
	store.page = PagedSessions{
		Items: []SessionSummary{{Session: Session{ID: 9, UserID: 7, AIAgentID: 1001, Title: "增长", Status: StatusActive}, Agent: store.agent}},
		Page:  1, PageSize: 20, Total: 1,
	}
	svc := NewService(store, &fakeModel{reply: "unused"})

	got, err := svc.List(context.Background(), 7, ListOptions{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if got.Total != 1 || len(got.Items) != 1 || store.listUserID != 7 || store.listOptions.PageSize != 20 {
		t.Fatalf("paged sessions = %#v store=%#v", got, store)
	}
}

func noopEmit(StreamEvent) error { return nil }

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
	session     Session
	agent       Agent
	messages    []Message
	page        PagedSessions
	listUserID  int64
	listOptions ListOptions
	nextID      int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		session: Session{ID: 9, UserID: 7, AIAgentID: 1001, Title: "林理臣", Status: StatusActive},
		agent:   Agent{ID: 1001, Name: "林理臣", SystemPrompt: "你是林理臣。"},
		nextID:  10,
	}
}

func (s *fakeStore) ListSessions(_ context.Context, userID int64, opts ListOptions) (PagedSessions, error) {
	s.listUserID = userID
	s.listOptions = opts
	return s.page, nil
}

func (s *fakeStore) GetAgent(_ context.Context, id int64) (Agent, error) {
	if id != s.agent.ID {
		return Agent{}, ErrAgentNotFound
	}
	return s.agent, nil
}

func (s *fakeStore) GetSession(_ context.Context, userID, sessionID int64) (Session, error) {
	if s.session.ID != sessionID || s.session.UserID != userID || s.session.Status == StatusDeleted {
		return Session{}, ErrSessionNotFound
	}
	return s.session, nil
}

func (s *fakeStore) ListMessages(_ context.Context, userID, sessionID int64) ([]Message, error) {
	if s.session.UserID != userID || s.session.ID != sessionID {
		return nil, ErrSessionNotFound
	}
	return append([]Message(nil), s.messages...), nil
}

func (s *fakeStore) StartSend(_ context.Context, userID int64, in SendInput) (SendStart, error) {
	created := false
	if in.SessionID == 0 {
		s.session = Session{ID: s.nextID, UserID: userID, AIAgentID: in.AgentID, Title: makeSessionTitle(in.Content), Status: StatusActive, MessageCount: 2}
		s.nextID++
		created = true
	} else if in.SessionID != s.session.ID {
		return SendStart{}, ErrSessionNotFound
	}
	user := Message{ID: s.nextID, SessionID: s.session.ID, Role: RoleUser, SenderType: RoleUser, Content: in.Content, Status: MessageDone, SequenceNo: int64(len(s.messages) + 1), RequestID: &in.RequestID}
	s.nextID++
	ai := Message{ID: s.nextID, SessionID: s.session.ID, Role: RoleAssistant, SenderType: RoleAssistant, Status: MessageStreaming, SequenceNo: user.SequenceNo + 1}
	s.nextID++
	s.messages = append(s.messages, user, ai)
	return SendStart{Session: s.session, Agent: s.agent, UserMessage: user, AssistantMessage: ai, CreatedSession: created, History: []Message{user}}, nil
}

func (s *fakeStore) CompleteAssistant(_ context.Context, _ int64, messageID int64, content string) (Message, Session, error) {
	for i := range s.messages {
		if s.messages[i].ID == messageID {
			s.messages[i].Content = content
			s.messages[i].Status = MessageDone
			return s.messages[i], s.session, nil
		}
	}
	return Message{}, Session{}, ErrMessageNotFound
}

func (s *fakeStore) FailAssistant(_ context.Context, _ int64, messageID int64, status, errMsg string) (Message, error) {
	for i := range s.messages {
		if s.messages[i].ID == messageID {
			s.messages[i].Status = status
			s.messages[i].ErrorMessage = &errMsg
			return s.messages[i], nil
		}
	}
	return Message{}, ErrMessageNotFound
}

func (s *fakeStore) StartRetry(_ context.Context, _ int64, in RetryInput) (RetryStart, error) {
	for i := range s.messages {
		if s.messages[i].ID == in.MessageID && (s.messages[i].Status == MessageFailed || s.messages[i].Status == MessagePartial) {
			s.messages[i].Status = MessageStreaming
			s.messages[i].Content = ""
			s.messages[i].RequestID = &in.RequestID
			return RetryStart{Session: s.session, Agent: s.agent, AssistantMessage: s.messages[i], History: append([]Message(nil), s.messages...)}, nil
		}
	}
	return RetryStart{}, ErrMessageNotFound
}

func (s *fakeStore) DeleteSession(_ context.Context, userID, sessionID int64) error {
	if s.session.UserID != userID || s.session.ID != sessionID {
		return ErrSessionNotFound
	}
	s.session.Status = StatusDeleted
	return nil
}
