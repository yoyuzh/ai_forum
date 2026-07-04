package decision

import (
	"context"
	"math"
	"strings"
	"testing"

	"ai-forum/backend/internal/ai/modelclient"
)

func TestWillingnessScoreMatchesHandComputedFixture(t *testing.T) {
	tags := []TagScore{
		{Type: "topic", Weight: 0.8},
		{Type: "topic", Weight: 0.4},
		{Type: "intent", Weight: 0.5},
		{Type: "emotion", Weight: 0.2},
		{Type: "debate", Weight: 0.6},
		{Type: "risk", Weight: 0.3},
	}

	got := WillingnessScore(ScoreInput{
		Tags:             tags,
		ActivityScore:    0.7,
		RiskPenalty:      0.1,
		FrequencyPenalty: 0.05,
	})

	want := 0.8*0.7*0.35 + 0.6*0.3*0.35 +
		0.5*0.25 +
		0.2*0.15 +
		0.6*0.15 +
		0.7*0.10 -
		0.1 - 0.05
	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("score = %.4f, want %.4f", got, want)
	}
}

func TestWillingnessScoreCoefficients(t *testing.T) {
	tests := []struct {
		name string
		in   ScoreInput
		want float64
	}{
		{name: "topic", in: ScoreInput{Tags: []TagScore{{Type: "topic", Weight: 1}}}, want: 0.35},
		{name: "intent", in: ScoreInput{Tags: []TagScore{{Type: "intent", Weight: 1}}}, want: 0.25},
		{name: "emotion", in: ScoreInput{Tags: []TagScore{{Type: "emotion", Weight: 1}}}, want: 0.15},
		{name: "debate", in: ScoreInput{Tags: []TagScore{{Type: "debate", Weight: 1}}}, want: 0.15},
		{name: "activity", in: ScoreInput{ActivityScore: 1}, want: 0.10},
		{name: "risk penalty", in: ScoreInput{RiskPenalty: 0.2}, want: -0.2},
		{name: "frequency penalty", in: ScoreInput{FrequencyPenalty: 0.2}, want: -0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WillingnessScore(tt.in)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Fatalf("score = %.4f, want %.4f", got, tt.want)
			}
		})
	}
}

func TestSelectRepliesPicksTopScoresWhenAllBelowThreshold(t *testing.T) {
	agents := []AgentScore{
		{AgentID: 1, Score: 0.2, Threshold: 0.6},
		{AgentID: 2, Score: 0.3, Threshold: 0.6},
		{AgentID: 3, Score: 0.1, Threshold: 0.6, Fallback: true},
		{AgentID: 4, Score: 0.4, Threshold: 0.6},
	}

	selected := SelectReplies(agents)

	if len(selected) != 3 || selected[0].AgentID != 4 || selected[1].AgentID != 2 || selected[2].AgentID != 1 {
		t.Fatalf("selected = %#v, want top 3 scores [4, 2, 1]", selected)
	}
	for _, agent := range selected {
		if agent.Decision != DecisionReply {
			t.Fatalf("selected = %#v, want all REPLY", selected)
		}
	}
}

func TestSelectRepliesPicksTopThreeScores(t *testing.T) {
	agents := []AgentScore{
		{AgentID: 1, Score: 0.65, Threshold: 0.6},
		{AgentID: 2, Score: 0.9, Threshold: 0.6},
		{AgentID: 3, Score: 0.2, Threshold: 0.6, Fallback: true},
		{AgentID: 4, Score: 0.8, Threshold: 0.6},
	}

	selected := SelectReplies(agents)

	if len(selected) != 3 || selected[0].AgentID != 2 || selected[1].AgentID != 4 || selected[2].AgentID != 1 {
		t.Fatalf("selected = %#v, want top 3 [2, 4, 1]", selected)
	}
}

func TestHandlerWritesLogsAndEnqueuesSelectedReplies(t *testing.T) {
	agents := &recordingAgentReader{agents: []Agent{
		{ID: 1, ReplyThreshold: 0.6, ActivityLevel: 0.5, AllowAutoReply: true, Preferences: []Preference{{TagType: "topic", TagName: "ai", Weight: 1}}},
		{ID: 2, ReplyThreshold: 0.6, ActivityLevel: 0.5, AllowAutoReply: true, Fallback: true, Preferences: []Preference{{TagType: "topic", TagName: "general", Weight: 0.1}}},
	}}
	tags := &recordingTagReader{tags: []PostTag{{Type: "topic", Name: "ai"}}}
	logs := &recordingDecisionLogger{}
	enqueuer := &recordingReplyEnqueuer{}
	handler := NewHandler(agents, tags, logs, enqueuer)

	if err := handler.HandleDecideAIReply(context.Background(), 42); err != nil {
		t.Fatal(err)
	}

	if len(logs.logs) != 2 {
		t.Fatalf("decision logs = %d, want 2", len(logs.logs))
	}
	if len(enqueuer.agentIDs) != 2 || enqueuer.agentIDs[0] != 1 || enqueuer.agentIDs[1] != 2 {
		t.Fatalf("enqueued agents = %#v, want [1, 2]", enqueuer.agentIDs)
	}
}

func TestHandlerUsesModelWillingnessScores(t *testing.T) {
	agents := &recordingAgentReader{agents: []Agent{
		{ID: 1, Name: "local", Persona: "本地高分", ReplyThreshold: 0.6, ActivityLevel: 0.5, AllowAutoReply: true, Preferences: []Preference{{TagType: "topic", TagName: "ai", Weight: 1}}},
		{ID: 2, Name: "model", Persona: "模型高分", ReplyThreshold: 0.6, ActivityLevel: 0.5, AllowAutoReply: true, Preferences: []Preference{{TagType: "topic", TagName: "other", Weight: 0.1}}},
	}}
	tags := &recordingTagReader{tags: []PostTag{{Type: "topic", Name: "ai"}}}
	logs := &recordingDecisionLogger{}
	enqueuer := &recordingReplyEnqueuer{}
	scorer := recordingWillingnessScorer{scores: map[int64]float64{1: 0.2, 2: 0.95}}
	handler := NewHandler(agents, tags, logs, enqueuer)
	handler.SetWillingnessScorer(scorer)

	if err := handler.HandleDecideAIReply(context.Background(), 42); err != nil {
		t.Fatal(err)
	}

	if len(enqueuer.agentIDs) != 2 || enqueuer.agentIDs[0] != 2 || enqueuer.agentIDs[1] != 1 {
		t.Fatalf("enqueued agents = %#v, want [2, 1]", enqueuer.agentIDs)
	}
	for _, log := range logs.logs {
		if log.AgentID == 2 && math.Abs(log.WillingnessScore-0.95) > 0.0001 {
			t.Fatalf("model score was not logged: %#v", log)
		}
	}
}

func TestModelWillingnessScorerSendsPersonaAndParsesScore(t *testing.T) {
	client := &recordingModelClient{out: "```json\n{\"score\":0.82}\n```"}
	scorer := NewModelWillingnessScorer(client)

	score, err := scorer.ScoreWillingness(context.Background(), WillingnessInput{
		PostID:        42,
		Agent:         Agent{ID: 1001, Name: "林理臣", Persona: "你是林理臣，冷静分析。\n更多规则", ReplyThreshold: 0.58},
		Tags:          []PostTag{{Type: "topic", Name: "职业选择"}},
		FallbackScore: 0.5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if math.Abs(score-0.82) > 0.0001 {
		t.Fatalf("score = %.2f, want 0.82", score)
	}
	for _, want := range []string{"满分是1.00", "回复阈值是0.58", "id=1001 name=林理臣 persona=你是林理臣，冷静分析。", "topic=职业选择", "本地公式参考分：0.5000"} {
		if !strings.Contains(client.in.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, client.in.Prompt)
		}
	}
}

func TestHandlerSkipsRepliesForHighRiskTag(t *testing.T) {
	agents := &recordingAgentReader{agents: []Agent{
		{ID: 1, ReplyThreshold: 0.1, ActivityLevel: 1, AllowAutoReply: true, Preferences: []Preference{{TagType: "risk", TagName: "高风险", Weight: 1}}},
		{ID: 2, ReplyThreshold: 0.1, ActivityLevel: 1, AllowAutoReply: true, Fallback: true},
	}}
	tags := &recordingTagReader{tags: []PostTag{{Type: "risk", Name: "高风险"}}}
	logs := &recordingDecisionLogger{}
	enqueuer := &recordingReplyEnqueuer{}
	handler := NewHandler(agents, tags, logs, enqueuer)

	if err := handler.HandleDecideAIReply(context.Background(), 42); err != nil {
		t.Fatal(err)
	}

	if len(enqueuer.agentIDs) != 0 {
		t.Fatalf("enqueued agents = %#v, want none", enqueuer.agentIDs)
	}
	if len(logs.logs) != 2 {
		t.Fatalf("decision logs = %d, want 2", len(logs.logs))
	}
	for _, log := range logs.logs {
		if log.Decision != DecisionIgnore || log.Reason != "high risk tag" {
			t.Fatalf("log = %#v, want high risk ignore", log)
		}
	}
}

type recordingAgentReader struct {
	agents []Agent
}

func (r *recordingAgentReader) ListEnabledAgents(context.Context) ([]Agent, error) {
	return r.agents, nil
}

type recordingTagReader struct {
	tags []PostTag
}

func (r *recordingTagReader) ListPostTags(context.Context, int64) ([]PostTag, error) {
	return r.tags, nil
}

type recordingDecisionLogger struct {
	logs []Log
}

func (l *recordingDecisionLogger) WriteDecisionLog(_ context.Context, log Log) error {
	l.logs = append(l.logs, log)
	return nil
}

type recordingReplyEnqueuer struct {
	agentIDs []int64
}

func (e *recordingReplyEnqueuer) EnqueueAutoGenerateAIReply(_ context.Context, postID, agentID int64) error {
	e.agentIDs = append(e.agentIDs, agentID)
	return nil
}

type recordingWillingnessScorer struct {
	scores map[int64]float64
}

func (s recordingWillingnessScorer) ScoreWillingness(_ context.Context, in WillingnessInput) (float64, error) {
	return s.scores[in.Agent.ID], nil
}

type recordingModelClient struct {
	in  modelclient.Request
	out string
}

func (c *recordingModelClient) Generate(_ context.Context, in modelclient.Request) (string, error) {
	c.in = in
	return c.out, nil
}
