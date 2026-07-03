// Package moderation reviews unsafe content before it becomes visible.
package moderation

import (
	"context"
	"fmt"
	"strings"
)

type Input struct {
	Text string
}

type Result struct {
	Allowed bool
	Reason  string
	Tags    []string
}

type Reviewer interface {
	Review(context.Context, Input) (Result, error)
}

type RuleModerator struct {
	words []string
}

func NewRuleModerator(words []string) *RuleModerator {
	if len(words) == 0 {
		words = []string{"violence", "self-harm", "forbidden"}
	}
	return &RuleModerator{words: words}
}

func (m *RuleModerator) Review(_ context.Context, in Input) (Result, error) {
	text := strings.ToLower(in.Text)
	for _, word := range m.words {
		if word != "" && strings.Contains(text, strings.ToLower(word)) {
			return Result{Allowed: false, Reason: fmt.Sprintf("blocked word: %s", word), Tags: []string{"rule"}}, nil
		}
	}
	return Result{Allowed: true}, nil
}
