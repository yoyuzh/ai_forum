package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var requiredSnippets = map[string][]string{
	"internal/event/event.go": {
		`PostModerated    = "post.moderated"`,
		`AIReplyFailed    = "ai.reply.failed"`,
	},
	"internal/task/task.go": {
		`GenerateAIReply        = "generate_ai_reply"`,
		`SyncSearchIndex        = "sync_search_index"`,
		`SendNotification       = "send_notification"`,
		`RefreshHotScore        = "refresh_hot_score"`,
		`CleanupProcessedEvents = "cleanup_processed_events"`,
		`mux.HandleFunc(GenerateAIReply`,
		`mux.HandleFunc(SyncSearchIndex`,
		`mux.HandleFunc(SendNotification`,
	},
	"internal/bootstrap/bootstrap.go": {
		`search.NewSyncHandler`,
		`notification.NewHandler`,
		`NewObservedClient`,
	},
	"internal/notification/http_handler.go": {
		`func (h *HTTPHandler) MarkRead`,
		`func (h *HTTPHandler) MarkAllRead`,
	},
	"cmd/search-rebuild/main.go": {
		`RebuildAll(ctx)`,
	},
	"../e2e/tests/p13_notification_contract.spec.ts": {
		`mark-one and read-all`,
	},
	"../e2e/tests/p13_non_functional.spec.ts": {
		`real Playwright interaction timing`,
		`Cohere text/background pairs meet WCAG AA contrast`,
	},
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	if err := check(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check(root string) error {
	var missing []string
	for rel, snippets := range requiredSnippets {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			missing = append(missing, rel+": "+err.Error())
			continue
		}
		body := string(data)
		for _, snippet := range snippets {
			if !strings.Contains(body, snippet) {
				missing = append(missing, rel+": missing "+snippet)
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("P13 implementation completeness failed:\n%s", strings.Join(missing, "\n"))
	}
	return nil
}
