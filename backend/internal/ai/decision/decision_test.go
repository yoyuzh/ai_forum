package decision

import (
	"context"
	"math"
	"testing"
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

func TestSelectRepliesFallsBackWhenAllScoresLow(t *testing.T) {
	agents := []AgentScore{
		{AgentID: 1, Score: 0.2, Threshold: 0.6},
		{AgentID: 2, Score: 0.3, Threshold: 0.6},
		{AgentID: 3, Score: 0.1, Threshold: 0.6, Fallback: true},
	}

	selected := SelectReplies(agents)

	if len(selected) != 1 || selected[0].AgentID != 3 || selected[0].Decision != DecisionFallback {
		t.Fatalf("selected = %#v, want fallback agent 3", selected)
	}
}

func TestSelectRepliesIncludesAllOverThreshold(t *testing.T) {
	agents := []AgentScore{
		{AgentID: 1, Score: 0.7, Threshold: 0.6},
		{AgentID: 2, Score: 0.65, Threshold: 0.6},
		{AgentID: 3, Score: 0.2, Threshold: 0.6, Fallback: true},
	}

	selected := SelectReplies(agents)

	if len(selected) != 2 || selected[0].Decision != DecisionReply || selected[1].Decision != DecisionReply {
		t.Fatalf("selected = %#v, want two replies", selected)
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
	if len(enqueuer.agentIDs) != 1 || enqueuer.agentIDs[0] != 1 {
		t.Fatalf("enqueued agents = %#v, want [1]", enqueuer.agentIDs)
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
