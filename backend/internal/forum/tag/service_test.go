package tag

import "testing"

func TestGroupByType(t *testing.T) {
	grouped := GroupByType([]Tag{
		{Type: "topic", Name: "go"},
		{Type: "topic", Name: "backend"},
		{Type: "risk", Name: "spam"},
	})

	if len(grouped["topic"]) != 2 || grouped["risk"][0] != "spam" {
		t.Fatalf("grouped = %#v", grouped)
	}
}
