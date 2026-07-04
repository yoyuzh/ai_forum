package modelclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestBuildPromptIncludesPersonaAndPost(t *testing.T) {
	prompt := BuildPrompt(PromptInput{
		AgentName:     "cohere_observer",
		SystemPrompt:  "你是白总结。",
		PostTitle:     "Should AI reply?",
		PostContent:   "Discuss the tradeoffs.",
		ParentContent: "Parent context",
	})

	for _, want := range []string{"你是白总结。", "Should AI reply?", "Discuss the tradeoffs.", "Parent context"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
}

func TestObservedClientLogsModelCallFieldsAndRedactsPromptAndSecrets(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	client := NewObservedClient(
		fakeClient{out: "ok"},
		zap.New(core),
		"gpt-test",
	)

	_, err := client.Generate(context.Background(), Request{
		Prompt:        "secret prompt body",
		TaskID:        "task-7",
		TaskType:      "generate_ai_reply",
		PostID:        42,
		AIAgentID:     1001,
		TriggerType:   "AUTO",
		RetryCount:    2,
		APIKey:        "sk-secret",
		InternalToken: "internal-secret",
	})
	if err != nil {
		t.Fatal(err)
	}

	if logs.Len() != 1 {
		t.Fatalf("logs = %d, want 1", logs.Len())
	}
	entry := logs.All()[0]
	if entry.Message != "ai_model_call" {
		t.Fatalf("message = %q", entry.Message)
	}
	fields := entry.ContextMap()
	for _, key := range []string{"task_id", "task_type", "post_id", "ai_agent_id", "trigger_type", "model", "latency_ms", "status", "retry_count"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("missing log field %s in %#v", key, fields)
		}
	}
	if fields["status"] != "success" || fields["model"] != "gpt-test" {
		t.Fatalf("fields = %#v", fields)
	}
	logged := entry.Message + fieldsString(fields)
	for _, forbidden := range []string{"secret prompt body", "sk-secret", "internal-secret"} {
		if strings.Contains(logged, forbidden) {
			t.Fatalf("log leaked %q: %s", forbidden, logged)
		}
	}
}

func TestObservedClientLogsFailureStatusAndErrorMessage(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	client := NewObservedClient(fakeClient{err: errors.New("model down")}, zap.New(core), "gpt-test")

	if _, err := client.Generate(context.Background(), Request{TaskID: "task-8", TaskType: "generate_ai_reply"}); err == nil {
		t.Fatal("expected error")
	}

	fields := logs.All()[0].ContextMap()
	if fields["status"] != "error" || fields["error_message"] != "model down" {
		t.Fatalf("fields = %#v", fields)
	}
}

type fakeClient struct {
	out string
	err error
}

func (f fakeClient) Generate(context.Context, Request) (string, error) {
	return f.out, f.err
}

func fieldsString(fields map[string]any) string {
	var b strings.Builder
	for key, value := range fields {
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(fmt.Sprint(value))
	}
	return b.String()
}

func TestOpenAICompatibleClientPostsChatCompletion(t *testing.T) {
	var gotPath, gotAuth, gotModel string
	var gotMaxTokens int
	var gotTemperature float64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		var body struct {
			Model       string  `json:"model"`
			MaxTokens   int     `json:"max_tokens"`
			Temperature float64 `json:"temperature"`
			Messages    []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		gotModel = body.Model
		gotMaxTokens = body.MaxTokens
		gotTemperature = body.Temperature
		if len(body.Messages) != 2 || body.Messages[0].Role != "system" || body.Messages[0].Content != "sys" || body.Messages[1].Content != "hello" {
			t.Fatalf("messages = %#v", body.Messages)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"AI reply"}}]}`))
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(server.URL, "test-key", "gpt-test", server.Client())
	temp := 0.1
	got, err := client.Generate(context.Background(), Request{SystemPrompt: "sys", Prompt: "hello", MaxTokens: 300, Temperature: &temp})
	if err != nil {
		t.Fatal(err)
	}

	if got != "AI reply" {
		t.Fatalf("reply = %q", got)
	}
	if gotPath != "/v1/chat/completions" || gotAuth != "Bearer test-key" || gotModel != "gpt-test" {
		t.Fatalf("request path/auth/model = %s/%s/%s", gotPath, gotAuth, gotModel)
	}
	if gotMaxTokens != 300 || gotTemperature != 0.1 {
		t.Fatalf("params = %d/%f, want 300/0.1", gotMaxTokens, gotTemperature)
	}
}
