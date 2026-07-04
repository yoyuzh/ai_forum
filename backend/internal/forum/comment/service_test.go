package comment

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"ai-forum/backend/internal/outbox"
	"ai-forum/backend/internal/task"
)

func TestServiceCreateAndDeleteAppendOutbox(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{id: 9}
	hot := &recordingHotTracker{}
	var events []outbox.Event
	svc := NewService(repo, func(ctx context.Context, _ DBTX, event outbox.Event) error {
		events = append(events, event)
		return nil
	}, WithHotTracker(hot))

	c, err := svc.Create(context.Background(), tx, CreateInput{PostID: 42, UserID: 7, Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != 9 || repo.incremented {
		t.Fatalf("comment=%#v incremented=%v, want no hot-path MySQL counter update", c, repo.incremented)
	}
	if len(events) != 1 || events[0].EventType != "comment.created" {
		t.Fatalf("events after create = %#v", events)
	}

	if err := svc.Delete(context.Background(), tx, 42, 9); err != nil {
		t.Fatal(err)
	}
	if !repo.deleted || !repo.decremented {
		t.Fatalf("deleted=%v decremented=%v", repo.deleted, repo.decremented)
	}
	if len(events) != 2 || events[1].EventType != "comment.deleted" {
		t.Fatalf("events after delete = %#v", events)
	}
	if len(hot.deltas) != 1 || hot.deltas[0] != 1 {
		t.Fatalf("hot deltas = %#v, want +1 on create", hot.deltas)
	}
}

func TestServiceCreateMentionWritesMentionAndQueuesAfterCommit(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{id: 9, agents: map[string]MentionAgent{
		"cohere_observer": {ID: 1001, Name: "cohere_observer", Enabled: true, AllowMention: true},
	}}
	mentions := &recordingMentionLimiter{}
	queue := &recordingAfterCommitQueue{}
	svc := NewService(repo, noopAppend, WithMentionLimiter(mentions), WithAfterCommit(queue.AfterCommit), WithGenerateEnqueuer(queue))

	c, err := svc.Create(context.Background(), tx, CreateInput{PostID: 42, UserID: 7, Content: "hello @cohere_observer"})
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.mentions) != 1 || repo.mentions[0].CommentID != c.ID || repo.mentions[0].AIAgentID != 1001 {
		t.Fatalf("mentions = %#v", repo.mentions)
	}
	if len(queue.generate) != 0 {
		t.Fatalf("generate queued before commit = %#v", queue.generate)
	}

	queue.Run()

	if len(queue.generate) != 1 {
		t.Fatalf("generate after commit = %#v", queue.generate)
	}
	got := queue.generate[0]
	if got.PostID != 42 || got.AIAgentID != 1001 || got.TriggerType != "MENTION" || got.ParentCommentID == nil || *got.ParentCommentID != c.ID {
		t.Fatalf("generate payload = %#v", got)
	}
	if mentions.userID != 7 || mentions.count != 1 {
		t.Fatalf("rate limiter = user %d count %d, want 7/1", mentions.userID, mentions.count)
	}
}

func TestParseMentionNamesSupportsChineseAgentNames(t *testing.T) {
	got := parseMentionNames("请 @林理臣 和 @cohere_observer 看看")
	want := []string{"林理臣", "cohere_observer"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mentions = %#v, want %#v", got, want)
	}
}

func TestServiceCreateMentionRejectsOverThreeAIs(t *testing.T) {
	repo := &recordingRepository{id: 9, agents: map[string]MentionAgent{
		"a1": {ID: 1, Name: "a1", Enabled: true, AllowMention: true},
		"a2": {ID: 2, Name: "a2", Enabled: true, AllowMention: true},
		"a3": {ID: 3, Name: "a3", Enabled: true, AllowMention: true},
		"a4": {ID: 4, Name: "a4", Enabled: true, AllowMention: true},
	}}
	svc := NewService(repo, noopAppend)

	_, err := svc.Create(context.Background(), nil, CreateInput{PostID: 42, UserID: 7, Content: "@a1 @a2 @a3 @a4"})
	if err == nil {
		t.Fatal("expected over-limit error")
	}
}

func TestServiceCreateMentionSkipsDisabledAndDisallowedAgents(t *testing.T) {
	repo := &recordingRepository{id: 9, agents: map[string]MentionAgent{
		"enabled":    {ID: 1, Name: "enabled", Enabled: true, AllowMention: true},
		"disabled":   {ID: 2, Name: "disabled", Enabled: false, AllowMention: true},
		"disallowed": {ID: 3, Name: "disallowed", Enabled: true, AllowMention: false},
	}}
	queue := &recordingAfterCommitQueue{}
	svc := NewService(repo, noopAppend, WithAfterCommit(queue.AfterCommit), WithGenerateEnqueuer(queue))

	if _, err := svc.Create(context.Background(), nil, CreateInput{PostID: 42, UserID: 7, Content: "@enabled @disabled @disallowed"}); err != nil {
		t.Fatal(err)
	}
	queue.Run()

	if len(repo.mentions) != 1 || repo.mentions[0].AIAgentID != 1 || len(queue.generate) != 1 || queue.generate[0].AIAgentID != 1 {
		t.Fatalf("mentions=%#v generate=%#v", repo.mentions, queue.generate)
	}
}

func TestServiceCreateMentionRateLimitReturnsRateLimitError(t *testing.T) {
	repo := &recordingRepository{id: 9, agents: map[string]MentionAgent{
		"enabled": {ID: 1, Name: "enabled", Enabled: true, AllowMention: true},
	}}
	svc := NewService(repo, noopAppend, WithMentionLimiter(&recordingMentionLimiter{err: ErrMentionRateLimited}))

	_, err := svc.Create(context.Background(), nil, CreateInput{PostID: 42, UserID: 7, Content: "@enabled"})
	if !errors.Is(err, ErrMentionRateLimited) {
		t.Fatalf("err = %v, want ErrMentionRateLimited", err)
	}
}

func TestServiceCreateReplyToAIQueuesFollowupJudgeAfterCommit(t *testing.T) {
	parentID := int64(77)
	repo := &recordingRepository{id: 9, parent: &Comment{ID: parentID, PostID: 42, CommentType: "AI"}}
	queue := &recordingAfterCommitQueue{}
	svc := NewService(repo, noopAppend, WithAfterCommit(queue.AfterCommit), WithFollowupEnqueuer(queue))

	c, err := svc.Create(context.Background(), nil, CreateInput{PostID: 42, UserID: 7, ParentCommentID: &parentID, Content: "reply"})
	if err != nil {
		t.Fatal(err)
	}
	if len(queue.followup) != 0 {
		t.Fatalf("followup queued before commit = %#v", queue.followup)
	}
	queue.Run()

	if len(queue.followup) != 1 || queue.followup[0].PostID != 42 || queue.followup[0].ParentCommentID != parentID || queue.followup[0].ReplyCommentID != c.ID {
		t.Fatalf("followup after commit = %#v", queue.followup)
	}
}

func TestServiceCreateQueuesPostLevelFollowupJudgeForUserComment(t *testing.T) {
	repo := &recordingRepository{id: 9}
	queue := &recordingAfterCommitQueue{}
	svc := NewService(repo, noopAppend, WithAfterCommit(queue.AfterCommit), WithFollowupEnqueuer(queue))

	c, err := svc.Create(context.Background(), nil, CreateInput{PostID: 42, UserID: 7, Content: "new user comment"})
	if err != nil {
		t.Fatal(err)
	}
	queue.Run()
	if len(queue.followup) != 1 || queue.followup[0].PostID != 42 || queue.followup[0].ParentCommentID != 0 || queue.followup[0].ReplyCommentID != c.ID {
		t.Fatalf("followup = %#v, want post-level judge", queue.followup)
	}
}

func TestServiceCreateQueuesPostLevelFollowupForUserParent(t *testing.T) {
	parentID := int64(77)
	repo := &recordingRepository{id: 9, parent: &Comment{ID: parentID, PostID: 42, CommentType: "USER"}}
	queue := &recordingAfterCommitQueue{}
	svc := NewService(repo, noopAppend, WithAfterCommit(queue.AfterCommit), WithFollowupEnqueuer(queue))

	c, err := svc.Create(context.Background(), nil, CreateInput{PostID: 42, UserID: 7, ParentCommentID: &parentID, Content: "reply"})
	if err != nil {
		t.Fatal(err)
	}
	queue.Run()
	if len(queue.followup) != 1 || queue.followup[0].ParentCommentID != 0 || queue.followup[0].ReplyCommentID != c.ID {
		t.Fatalf("followup = %#v, want post-level judge", queue.followup)
	}
}

func noopAppend(context.Context, DBTX, outbox.Event) error { return nil }

type recordingRepository struct {
	id          int64
	incremented bool
	decremented bool
	deleted     bool
	agents      map[string]MentionAgent
	mentions    []CommentMention
	parent      *Comment
}

func (r *recordingRepository) Create(_ context.Context, _ DBTX, c Comment) (Comment, error) {
	c.ID = r.id
	return c, nil
}

func (r *recordingRepository) IncrementCommentCount(context.Context, DBTX, int64) error {
	r.incremented = true
	return nil
}

func (r *recordingRepository) SoftDelete(context.Context, DBTX, int64) error {
	r.deleted = true
	return nil
}

func (r *recordingRepository) DecrementCommentCount(context.Context, DBTX, int64) error {
	r.decremented = true
	return nil
}

func (r *recordingRepository) FindMentionAgents(_ context.Context, _ DBTX, names []string) ([]MentionAgent, error) {
	var out []MentionAgent
	for _, name := range names {
		if agent, ok := r.agents[name]; ok {
			out = append(out, agent)
		}
	}
	return out, nil
}

func (r *recordingRepository) CreateMention(_ context.Context, _ DBTX, mention CommentMention) error {
	r.mentions = append(r.mentions, mention)
	return nil
}

func (r *recordingRepository) Get(_ context.Context, _ DBTX, id int64) (Comment, error) {
	if r.parent != nil && r.parent.ID == id {
		return *r.parent, nil
	}
	return Comment{}, ErrCommentNotFound
}

func (r *recordingRepository) ListByPost(context.Context, DBTX, int64) ([]Comment, error) {
	return nil, nil
}

type recordingMentionLimiter struct {
	userID int64
	count  int
	err    error
}

func (l *recordingMentionLimiter) AllowMentions(_ context.Context, userID int64, count int) error {
	l.userID = userID
	l.count = count
	return l.err
}

type recordingAfterCommitQueue struct {
	callbacks []func(context.Context) error
	generate  []task.GenerateAIReplyPayload
	followup  []task.JudgeAIFollowupPayload
}

func (q *recordingAfterCommitQueue) AfterCommit(fn func(context.Context) error) {
	q.callbacks = append(q.callbacks, fn)
}

func (q *recordingAfterCommitQueue) Run() {
	for _, fn := range q.callbacks {
		_ = fn(context.Background())
	}
}

func (q *recordingAfterCommitQueue) EnqueueGenerateAIReply(_ context.Context, payload task.GenerateAIReplyPayload) error {
	q.generate = append(q.generate, payload)
	return nil
}

func (q *recordingAfterCommitQueue) EnqueueJudgeAIFollowup(_ context.Context, payload task.JudgeAIFollowupPayload) error {
	q.followup = append(q.followup, payload)
	return nil
}

type recordingHotTracker struct {
	deltas []int64
}

func (h *recordingHotTracker) RecordInteraction(_ context.Context, postID int64, counter HotCounter, delta int64) error {
	if postID != 42 || counter != HotCounterComment {
		return nil
	}
	h.deltas = append(h.deltas, delta)
	return nil
}
