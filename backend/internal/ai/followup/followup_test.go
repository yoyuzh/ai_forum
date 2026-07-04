package followup

import (
	"context"
	"errors"
	"testing"

	"ai-forum/backend/internal/task"
)

func TestHandlerEnqueuesFollowupWhenModelReturnsTrue(t *testing.T) {
	repo := &recordingRepository{
		post:   Post{ID: 42, Title: "title", Content: "body"},
		parent: Comment{ID: 77, PostID: 42, CommentType: "AI", AIAgentID: 1001, Content: "ai"},
		reply:  Comment{ID: 88, PostID: 42, CommentType: "USER", UserID: 7, Content: "user"},
	}
	enqueuer := &recordingGenerateEnqueuer{}
	h := NewHandler(repo, model{out: `{"should_reply":true,"reason":"asked a direct question"}`}, enqueuer)

	if err := h.HandleJudgeAIFollowup(context.Background(), task.JudgeAIFollowupPayload{PostID: 42, ParentCommentID: 77, ReplyCommentID: 88}); err != nil {
		t.Fatal(err)
	}

	if len(enqueuer.payloads) != 1 {
		t.Fatalf("payloads = %#v", enqueuer.payloads)
	}
	got := enqueuer.payloads[0]
	if got.PostID != 42 || got.ParentCommentID == nil || *got.ParentCommentID != 88 || got.AIAgentID != 1001 || got.TriggerType != "FOLLOWUP" {
		t.Fatalf("payload = %#v", got)
	}
}

func TestHandlerDefaultsFalseOnAnomalies(t *testing.T) {
	cases := []struct {
		name string
		out  string
		err  error
	}{
		{name: "call failure", err: errors.New("timeout")},
		{name: "non json", out: "yes"},
		{name: "missing should_reply", out: `{"reason":"x"}`},
		{name: "non boolean", out: `{"should_reply":"yes","reason":"x"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &recordingRepository{
				post:   Post{ID: 42},
				parent: Comment{ID: 77, PostID: 42, CommentType: "AI", AIAgentID: 1001},
				reply:  Comment{ID: 88, PostID: 42, CommentType: "USER", UserID: 7},
			}
			enqueuer := &recordingGenerateEnqueuer{}
			h := NewHandler(repo, model{out: tc.out, err: tc.err}, enqueuer)

			if err := h.HandleJudgeAIFollowup(context.Background(), task.JudgeAIFollowupPayload{PostID: 42, ParentCommentID: 77, ReplyCommentID: 88}); err != nil {
				t.Fatal(err)
			}
			if len(enqueuer.payloads) != 0 {
				t.Fatalf("payloads = %#v, want none", enqueuer.payloads)
			}
		})
	}
}

func TestHandlerSelectsAIsForPostLevelUserComment(t *testing.T) {
	repo := &recordingRepository{
		post:       Post{ID: 42, Title: "title", Content: "body"},
		reply:      Comment{ID: 88, PostID: 42, CommentType: "USER", UserID: 7, Content: "anyone?"},
		candidates: []Candidate{{AIAgentID: 1001, Name: "a", Content: "first"}, {AIAgentID: 1002, Name: "b", Content: "second"}},
	}
	enqueuer := &recordingGenerateEnqueuer{}
	h := NewHandler(repo, model{out: `{"agent_ids":[1002,9999]}`}, enqueuer)

	if err := h.HandleJudgeAIFollowup(context.Background(), task.JudgeAIFollowupPayload{PostID: 42, ReplyCommentID: 88}); err != nil {
		t.Fatal(err)
	}

	if len(enqueuer.payloads) != 1 {
		t.Fatalf("payloads = %#v, want one selected candidate", enqueuer.payloads)
	}
	got := enqueuer.payloads[0]
	if got.PostID != 42 || got.ParentCommentID == nil || *got.ParentCommentID != 88 || got.AIAgentID != 1002 || got.TriggerType != "FOLLOWUP" {
		t.Fatalf("payload = %#v", got)
	}
}

func TestHandlerDoesNotEnqueueWhenPostLevelModelReturnsEmpty(t *testing.T) {
	for _, out := range []string{`{"agent_ids":[]}`, `{"agent_id":null}`} {
		t.Run(out, func(t *testing.T) {
			repo := &recordingRepository{
				post:       Post{ID: 42},
				reply:      Comment{ID: 88, PostID: 42, CommentType: "USER", UserID: 7},
				candidates: []Candidate{{AIAgentID: 1001, Name: "a", Content: "first"}},
			}
			enqueuer := &recordingGenerateEnqueuer{}
			h := NewHandler(repo, model{out: out}, enqueuer)

			if err := h.HandleJudgeAIFollowup(context.Background(), task.JudgeAIFollowupPayload{PostID: 42, ReplyCommentID: 88}); err != nil {
				t.Fatal(err)
			}
			if len(enqueuer.payloads) != 0 {
				t.Fatalf("payloads = %#v, want none", enqueuer.payloads)
			}
		})
	}
}

func TestHandlerDoesNotEnqueueForAIReplyOrFollowupCap(t *testing.T) {
	cases := []struct {
		name  string
		reply Comment
		count int
	}{
		{name: "ai to ai", reply: Comment{ID: 88, PostID: 42, CommentType: "AI", AIAgentID: 1002}},
		{name: "cap reached", reply: Comment{ID: 88, PostID: 42, CommentType: "USER", UserID: 7}, count: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &recordingRepository{
				post:          Post{ID: 42},
				parent:        Comment{ID: 77, PostID: 42, CommentType: "AI", AIAgentID: 1001},
				reply:         tc.reply,
				followupCount: tc.count,
			}
			enqueuer := &recordingGenerateEnqueuer{}
			h := NewHandler(repo, model{out: `{"should_reply":true,"reason":"x"}`}, enqueuer)

			if err := h.HandleJudgeAIFollowup(context.Background(), task.JudgeAIFollowupPayload{PostID: 42, ParentCommentID: 77, ReplyCommentID: 88}); err != nil {
				t.Fatal(err)
			}
			if len(enqueuer.payloads) != 0 {
				t.Fatalf("payloads = %#v, want none", enqueuer.payloads)
			}
		})
	}
}

type recordingRepository struct {
	post          Post
	parent        Comment
	reply         Comment
	candidates    []Candidate
	followupCount int
}

func (r *recordingRepository) LoadPost(_ context.Context, id int64) (Post, error) {
	if id == r.post.ID {
		return r.post, nil
	}
	return Post{}, errors.New("post not found")
}

func (r *recordingRepository) LoadComment(_ context.Context, id int64) (Comment, error) {
	switch id {
	case r.parent.ID:
		return r.parent, nil
	case r.reply.ID:
		return r.reply, nil
	default:
		return Comment{}, errors.New("comment not found")
	}
}

func (r *recordingRepository) CountFollowups(_ context.Context, postID, agentID int64) (int, error) {
	return r.followupCount, nil
}

func (r *recordingRepository) ListPostAICandidates(context.Context, int64) ([]Candidate, error) {
	return r.candidates, nil
}

type model struct {
	out string
	err error
}

func (m model) Generate(_ context.Context, _ Prompt) (string, error) {
	return m.out, m.err
}

type recordingGenerateEnqueuer struct {
	payloads []task.GenerateAIReplyPayload
}

func (e *recordingGenerateEnqueuer) EnqueueGenerateAIReply(_ context.Context, payload task.GenerateAIReplyPayload) error {
	e.payloads = append(e.payloads, payload)
	return nil
}
