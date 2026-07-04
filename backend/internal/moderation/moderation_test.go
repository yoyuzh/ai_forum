package moderation

import (
	"context"
	"testing"
)

func TestRuleModeratorBlocksSensitiveWords(t *testing.T) {
	result, err := NewRuleModerator([]string{"forbidden"}).Review(context.Background(), Input{Text: "a forbidden reply"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Allowed || result.Reason == "" {
		t.Fatalf("result = %#v, want blocked with reason", result)
	}
}

func TestRuleModeratorAllowsCleanText(t *testing.T) {
	result, err := NewRuleModerator(nil).Review(context.Background(), Input{Text: "A useful answer."})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Allowed {
		t.Fatalf("result = %#v, want allowed", result)
	}
}
