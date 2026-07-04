package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

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

func TestRegisterHandlersIncludesGenerateAIReply(t *testing.T) {
	db := &taskDBTX{}
	reply := &recordingGenerateAIReplyHandler{}
	mux := asynq.NewServeMux()
	RegisterHandlers(mux, db, Handlers{GenerateAIReply: reply.HandleGenerateAIReply})
	parentID := int64(77)
	payload, err := json.Marshal(GenerateAIReplyPayload{PostID: 42, ParentCommentID: &parentID, AIAgentID: 1001, TriggerType: "MENTION"})
	if err != nil {
		t.Fatal(err)
	}

	if err := mux.ProcessTask(context.Background(), asynq.NewTask(GenerateAIReply, payload)); err != nil {
		t.Fatal(err)
	}

	if reply.postID != 42 || reply.parentID == nil || *reply.parentID != 77 || reply.agentID != 1001 || reply.triggerType != "MENTION" {
		t.Fatalf("reply task = post %d parent %v agent %d trigger %q, want 42/77/1001/MENTION", reply.postID, reply.parentID, reply.agentID, reply.triggerType)
	}
}

func TestRegisterHandlersIncludesJudgeAIFollowup(t *testing.T) {
	db := &taskDBTX{}
	followup := &recordingJudgeAIFollowupHandler{}
	mux := asynq.NewServeMux()
	RegisterHandlers(mux, db, Handlers{JudgeAIFollowup: followup.HandleJudgeAIFollowup})
	payload, err := json.Marshal(JudgeAIFollowupPayload{PostID: 42, ParentCommentID: 77, ReplyCommentID: 88})
	if err != nil {
		t.Fatal(err)
	}

	if err := mux.ProcessTask(context.Background(), asynq.NewTask(JudgeAIFollowup, payload)); err != nil {
		t.Fatal(err)
	}

	if followup.postID != 42 || followup.parentID != 77 || followup.replyID != 88 {
		t.Fatalf("followup task = post %d parent %d reply %d, want 42/77/88", followup.postID, followup.parentID, followup.replyID)
	}
}

func TestRegisterHandlersIncludesRefreshHotScore(t *testing.T) {
	db := &taskDBTX{}
	hot := &recordingRefreshHotScoreHandler{}
	mux := asynq.NewServeMux()
	RegisterHandlers(mux, db, Handlers{RefreshHotScore: hot.HandleRefreshHotScore})

	if err := mux.ProcessTask(context.Background(), asynq.NewTask(RefreshHotScore, nil)); err != nil {
		t.Fatal(err)
	}
	if hot.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", hot.calls)
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
	parentID := int64(77)

	if err := reply.EnqueueGenerateAIReply(context.Background(), GenerateAIReplyPayload{PostID: 42, ParentCommentID: &parentID, AIAgentID: 1001, TriggerType: "MENTION"}); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskType != GenerateAIReply || enqueuer.postID != 42 || enqueuer.parentID == nil || *enqueuer.parentID != 77 || enqueuer.agentID != 1001 || enqueuer.triggerType != "MENTION" {
		t.Fatalf("enqueued = %s post=%d parent=%v agent=%d trigger=%q, want generate_ai_reply post=42 parent=77 agent=1001 trigger=MENTION", enqueuer.taskType, enqueuer.postID, enqueuer.parentID, enqueuer.agentID, enqueuer.triggerType)
	}
	if enqueuer.taskID != "generate_ai_reply:42:77:1001:MENTION" {
		t.Fatalf("task id = %q, want deterministic generate_ai_reply:42:77:1001:MENTION", enqueuer.taskID)
	}
	if enqueuer.maxRetry != GenerateAIReplyMaxRetries {
		t.Fatalf("max retry = %d, want %d", enqueuer.maxRetry, GenerateAIReplyMaxRetries)
	}
}

func TestGenerateAIReplyRetryUsesDistinctTaskID(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	reply := NewGenerateAIReplyEnqueuer(enqueuer)

	if err := reply.EnqueueGenerateAIReplyRetry(context.Background(), GenerateAIReplyPayload{PostID: 42, AIAgentID: 1001, TriggerType: "AUTO"}, "7:2"); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskID != "generate_ai_reply:42:0:1001:AUTO:retry:7:2" {
		t.Fatalf("task id = %q, want retry suffix", enqueuer.taskID)
	}
}

func TestGenerateAIReplyRetryDelayIsTenMinutes(t *testing.T) {
	got := retryDelay(1, errors.New("boom"), asynq.NewTask(GenerateAIReply, nil))
	if got != 10*time.Minute {
		t.Fatalf("retry delay = %s, want 10m", got)
	}
}

func TestJudgeAIFollowupEnqueuerUsesTaskContract(t *testing.T) {
	enqueuer := &recordingEnqueuer{}
	followup := NewJudgeAIFollowupEnqueuer(enqueuer)

	if err := followup.EnqueueJudgeAIFollowup(context.Background(), JudgeAIFollowupPayload{PostID: 42, ParentCommentID: 77, ReplyCommentID: 88}); err != nil {
		t.Fatal(err)
	}

	if enqueuer.taskType != JudgeAIFollowup || enqueuer.postID != 42 || enqueuer.parentCommentID != 77 || enqueuer.replyCommentID != 88 {
		t.Fatalf("enqueued = %s post=%d parent=%d reply=%d, want judge_ai_followup 42/77/88", enqueuer.taskType, enqueuer.postID, enqueuer.parentCommentID, enqueuer.replyCommentID)
	}
	if enqueuer.taskID != "judge_ai_followup:42:77:88" {
		t.Fatalf("task id = %q, want deterministic judge_ai_followup:42:77:88", enqueuer.taskID)
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

type recordingGenerateAIReplyHandler struct {
	postID      int64
	parentID    *int64
	agentID     int64
	triggerType string
}

func (h *recordingGenerateAIReplyHandler) HandleGenerateAIReply(_ context.Context, payload GenerateAIReplyPayload) error {
	h.postID = payload.PostID
	h.parentID = payload.ParentCommentID
	h.agentID = payload.AIAgentID
	h.triggerType = payload.TriggerType
	return nil
}

type recordingJudgeAIFollowupHandler struct {
	postID   int64
	parentID int64
	replyID  int64
}

func (h *recordingJudgeAIFollowupHandler) HandleJudgeAIFollowup(_ context.Context, payload JudgeAIFollowupPayload) error {
	h.postID = payload.PostID
	h.parentID = payload.ParentCommentID
	h.replyID = payload.ReplyCommentID
	return nil
}

type recordingRefreshHotScoreHandler struct {
	calls int
}

func (h *recordingRefreshHotScoreHandler) HandleRefreshHotScore(context.Context) error {
	h.calls++
	return nil
}
