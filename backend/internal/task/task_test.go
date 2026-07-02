package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/hibiken/asynq"
)

func TestAllTaskTypesIncludeCleanup(t *testing.T) {
	want := []string{
		"tag_post",
		"decide_ai_reply",
		"generate_ai_reply",
		"judge_ai_followup",
		"moderate_ai_reply",
		"sync_search_index",
		"send_notification",
		"refresh_hot_score",
		"cleanup_processed_events",
	}
	got := Types()
	if len(got) != len(want) {
		t.Fatalf("task type count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("task type[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCleanupProcessedEventsDeletesOnlyOldRows(t *testing.T) {
	db := &taskDBTX{}

	if err := CleanupProcessedEventsRows(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.query, "DELETE FROM processed_events") {
		t.Fatalf("query = %q", db.query)
	}
	if !strings.Contains(db.query, "INTERVAL 30 DAY") {
		t.Fatalf("query must keep recent rows: %q", db.query)
	}
}

func TestCronContractsHaveHandlerOwners(t *testing.T) {
	for _, cronType := range CronTypes() {
		contract, ok := CronContract(cronType)
		if !ok {
			t.Fatalf("missing cron contract for %s", cronType)
		}
		if contract.HandlerOwner == "" {
			t.Fatalf("missing handler owner for %s", cronType)
		}
	}
}

func TestRegisterHandlersIncludesCleanupAndP6TagPost(t *testing.T) {
	db := &taskDBTX{}
	tagging := &recordingTagPostHandler{}
	mux := asynq.NewServeMux()
	RegisterHandlers(mux, db, Handlers{TagPost: tagging.HandleTagPost})

	err := mux.ProcessTask(context.Background(), asynq.NewTask(CleanupProcessedEvents, nil))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(db.query, "DELETE FROM processed_events") {
		t.Fatalf("cleanup handler query = %q", db.query)
	}

	payload, err := json.Marshal(TagPostPayload{PostID: 42})
	if err != nil {
		t.Fatal(err)
	}
	if err := mux.ProcessTask(context.Background(), asynq.NewTask(TagPost, payload)); err != nil {
		t.Fatal(err)
	}
	if tagging.postID != 42 {
		t.Fatalf("tag post id = %d, want 42", tagging.postID)
	}
}

func TestRegisterHandlersDedupsAsynqTaskByTaskID(t *testing.T) {
	db := &taskDBTX{}
	tagging := &recordingTagPostHandler{}
	mux := asynq.NewServeMux()
	RegisterHandlers(mux, db, Handlers{TagPost: tagging.HandleTagPost})
	payload, err := json.Marshal(TagPostPayload{PostID: 42})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if err := mux.ProcessTask(ctx, asynq.NewTask(TagPost, payload)); err != nil {
		t.Fatal(err)
	}
	if err := mux.ProcessTask(ctx, asynq.NewTask(TagPost, payload)); err != nil {
		t.Fatal(err)
	}

	if tagging.calls != 1 {
		t.Fatalf("tag_post calls = %d, want 1", tagging.calls)
	}
}

func TestGenerateAIReplyEnqueuerUsesTaskContract(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	reply := NewGenerateAIReplyEnqueuer(enqueuer)

	if err := reply.EnqueueGenerateAIReply(context.Background(), 42, 1001); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskType != GenerateAIReply || enqueuer.postID != 42 || enqueuer.agentID != 1001 {
		t.Fatalf("enqueued = %s post=%d agent=%d, want generate_ai_reply post=42 agent=1001", enqueuer.taskType, enqueuer.postID, enqueuer.agentID)
	}
	if enqueuer.taskID != "generate_ai_reply:42:1001" {
		t.Fatalf("task id = %q, want deterministic generate_ai_reply:42:1001", enqueuer.taskID)
	}
}

type taskDBTX struct {
	query     string
	processed map[string]bool
}

func (d *taskDBTX) ExecContext(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	d.query = query
	if strings.Contains(query, "INSERT INTO processed_events") {
		if d.processed == nil {
			d.processed = map[string]bool{}
		}
		if len(args) >= 2 {
			d.processed[args[0].(string)+"/"+args[1].(string)] = true
		}
	}
	return fakeTaskResult(1), nil
}

func (d *taskDBTX) GetContext(_ context.Context, dest interface{}, _ string, args ...interface{}) error {
	if len(args) >= 2 {
		key := args[0].(string) + "/" + args[1].(string)
		if d.processed[key] {
			*(dest.(*int)) = 1
			return nil
		}
		*(dest.(*int)) = 0
		return nil
	}
	return errors.New("unexpected get")
}

func (d *taskDBTX) SelectContext(context.Context, interface{}, string, ...interface{}) error {
	return errors.New("unexpected select")
}

type fakeTaskResult int64

func (r fakeTaskResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeTaskResult) RowsAffected() (int64, error) { return int64(r), nil }

type recordingTagPostHandler struct {
	postID int64
	calls  int
}

func (h *recordingTagPostHandler) HandleTagPost(_ context.Context, postID int64) error {
	h.calls++
	h.postID = postID
	return nil
}
