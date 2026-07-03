package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-forum/backend/internal/auth"
	"ai-forum/backend/internal/rbac"
)

func TestHandlerListsDecisionLogsWithExplainabilityFields(t *testing.T) {
	h := NewHandler(&fakeStore{}, mustAuthorizer(t))
	req := httptest.NewRequest(http.MethodGet, "/api/admin/decision-logs", nil)
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 1, Role: "ADMIN"}))
	rec := httptest.NewRecorder()

	h.ListDecisionLogs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var rows []DecisionLog
	if err := json.NewDecoder(rec.Body).Decode(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	got := rows[0]
	if got.AIAgentName != "cohere_observer" || got.Decision != "FALLBACK" || !got.Fallback {
		t.Fatalf("missing decision explainability fields: %#v", got)
	}
	if len(got.HitTags) != 1 || got.HitTags[0] != "topic:general" {
		t.Fatalf("hitTags = %#v", got.HitTags)
	}
}

func TestHandlerUpdatesAgentAndDeniesRetryByRBAC(t *testing.T) {
	store := &fakeStore{}
	h := NewHandler(store, mustAuthorizer(t))

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/ai-agents/1001", strings.NewReader(`{"replyThreshold":0.72,"allowAutoReply":false}`))
	req.SetPathValue("agentId", "1001")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 1, Role: "ADMIN"}))
	rec := httptest.NewRecorder()
	h.UpdateAgent(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if store.updatedAgentID != 1001 || store.updatedAgent.ReplyThreshold == nil || *store.updatedAgent.ReplyThreshold != 0.72 {
		t.Fatalf("update not forwarded: id=%d update=%#v", store.updatedAgentID, store.updatedAgent)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/ai-tasks/55/retry", nil)
	req.SetPathValue("taskId", "55")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 2, Role: "USER"}))
	rec = httptest.NewRecorder()
	h.RetryTask(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("retry status = %d, want 403", rec.Code)
	}
}

type fakeStore struct {
	updatedAgentID int64
	updatedAgent   AgentUpdate
}

func (f *fakeStore) ListUsers(context.Context) ([]User, error)              { return nil, nil }
func (f *fakeStore) ListPosts(context.Context) ([]Post, error)              { return nil, nil }
func (f *fakeStore) ListComments(context.Context) ([]Comment, error)        { return nil, nil }
func (f *fakeStore) ListAgents(context.Context) ([]Agent, error)            { return nil, nil }
func (f *fakeStore) ListTasks(context.Context) ([]Task, error)              { return nil, nil }
func (f *fakeStore) ListTags(context.Context) ([]Tag, error)                { return nil, nil }
func (f *fakeStore) ListPreferences(context.Context) ([]Preference, error)  { return nil, nil }
func (f *fakeStore) RetryTask(context.Context, int64) (Task, error)         { return Task{}, nil }
func (f *fakeStore) TerminateTask(context.Context, int64) (Task, error)     { return Task{}, nil }
func (f *fakeStore) MarkTaskProcessed(context.Context, int64) (Task, error) { return Task{}, nil }

func (f *fakeStore) ListDecisionLogs(context.Context) ([]DecisionLog, error) {
	return []DecisionLog{{
		ID: 7, PostID: 42, AIAgentID: 1001, AIAgentName: "cohere_observer",
		TriggerType: "AUTO", WillingnessScore: 0.32, ThresholdValue: 0.6,
		Decision: "FALLBACK", Reason: "fallback-invoked", Fallback: true,
		HitTags: []string{"topic:general"}, TaskID: ptrInt64(55),
	}}, nil
}

func (f *fakeStore) UpdateAgent(_ context.Context, id int64, update AgentUpdate) (Agent, error) {
	f.updatedAgentID = id
	f.updatedAgent = update
	return Agent{ID: id, Name: "cohere_observer"}, nil
}

func mustAuthorizer(t *testing.T) *rbac.Authorizer {
	t.Helper()
	authz, err := rbac.NewAuthorizer(rbac.DefaultModelPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := authz.SeedAdminPolicies(); err != nil {
		t.Fatal(err)
	}
	return authz
}

func ptrInt64(v int64) *int64 { return &v }
