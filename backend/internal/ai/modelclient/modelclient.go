// Package modelclient owns low-level AI provider calls.
package modelclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Request struct {
	Prompt        string
	TaskID        string
	TaskType      string
	PostID        int64
	AIAgentID     int64
	TriggerType   string
	RetryCount    int
	APIKey        string
	InternalToken string
}

type Client interface {
	Generate(context.Context, Request) (string, error)
}

type InfoLogger interface {
	Info(string, ...zap.Field)
}

type ObservedClient struct {
	next  Client
	log   InfoLogger
	model string
}

func NewObservedClient(next Client, log InfoLogger, model string) *ObservedClient {
	if log == nil {
		log = zap.NewNop()
	}
	return &ObservedClient{next: next, log: log, model: model}
}

func (c *ObservedClient) Generate(ctx context.Context, in Request) (string, error) {
	start := time.Now()
	out, err := c.next.Generate(ctx, in)
	fields := []zap.Field{
		zap.String("task_id", in.TaskID),
		zap.String("task_type", in.TaskType),
		zap.Int64("post_id", in.PostID),
		zap.Int64("ai_agent_id", in.AIAgentID),
		zap.String("trigger_type", in.TriggerType),
		zap.String("model", c.model),
		zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		zap.Int("retry_count", in.RetryCount),
	}
	if err != nil {
		fields = append(fields, zap.String("status", "error"), zap.String("error_message", err.Error()))
	} else {
		fields = append(fields, zap.String("status", "success"))
	}
	c.log.Info("ai_model_call", fields...)
	return out, err
}

type PromptInput struct {
	AgentName     string `db:"agent_name"`
	PostTitle     string `db:"post_title"`
	PostContent   string `db:"post_content"`
	ParentContent string `db:"parent_content"`
}

func BuildPrompt(in PromptInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are %s. Reply as a concise forum participant.\n\nPost: %s\n%s", in.AgentName, in.PostTitle, in.PostContent)
	if strings.TrimSpace(in.ParentContent) != "" {
		fmt.Fprintf(&b, "\n\nParent comment: %s", in.ParentContent)
	}
	return b.String()
}

type OpenAICompatibleClient struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

func NewOpenAICompatibleClient(baseURL, apiKey, model string, httpClient *http.Client) *OpenAICompatibleClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &OpenAICompatibleClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		http:    httpClient,
	}
}

func (c *OpenAICompatibleClient) Generate(ctx context.Context, in Request) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model": c.model,
		"messages": []map[string]string{{
			"role":    "user",
			"content": in.Prompt,
		}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("model status %d", resp.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("model returned no choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
