package comment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-forum/backend/internal/auth"
)

func TestHandlerCreateUsesAuthenticatedSubjectAndPostID(t *testing.T) {
	svc := &recordingCommentService{}
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error { return fn(nil) })
	req := httptest.NewRequest(http.MethodPost, "/api/posts/42/comments", strings.NewReader(`{"content":"hello"}`))
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7}))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if svc.input.PostID != 42 || svc.input.UserID != 7 || svc.input.Content != "hello" {
		t.Fatalf("input = %#v", svc.input)
	}
}

func TestHandlerCreatePassesParentCommentIDAndRunsAfterCommit(t *testing.T) {
	svc := &recordingCommentService{}
	var callbacks []func(context.Context) error
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error { return fn(nil) }, WithHandlerAfterCommit(func(fn func(context.Context) error) {
		callbacks = append(callbacks, fn)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/posts/42/comments", strings.NewReader(`{"content":"hello","parent_comment_id":77}`))
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7}))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if svc.input.ParentCommentID == nil || *svc.input.ParentCommentID != 77 {
		t.Fatalf("parent = %v, want 77", svc.input.ParentCommentID)
	}
	if len(callbacks) != 1 {
		t.Fatalf("callbacks = %d, want 1", len(callbacks))
	}
}

func TestHandlerCreateMapsMentionRateLimitTo429(t *testing.T) {
	svc := &recordingCommentService{err: ErrMentionRateLimited}
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error { return fn(nil) })
	req := httptest.NewRequest(http.MethodPost, "/api/posts/42/comments", strings.NewReader(`{"content":"@enabled"}`))
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7}))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429; body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerListUsesPostID(t *testing.T) {
	svc := &recordingCommentService{
		listed: []Comment{{
			ID:          8,
			PostID:      42,
			UserID:      7,
			Content:     "hello",
			CommentType: "USER",
			Author:      &Author{Username: "yoyuzh", IsAI: false},
		}},
	}
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error { return fn(nil) })
	req := httptest.NewRequest(http.MethodGet, "/api/posts/42/comments", nil)
	req.SetPathValue("postId", "42")
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if svc.listPostID != 42 {
		t.Fatalf("list postID = %d, want 42", svc.listPostID)
	}
	if !strings.Contains(rec.Body.String(), `"id":8`) {
		t.Fatalf("body = %q, want listed comment", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"username":"yoyuzh"`) {
		t.Fatalf("body = %q, want author nickname", rec.Body.String())
	}
}

type recordingCommentService struct {
	input      CreateInput
	listPostID int64
	listed     []Comment
	err        error
}

func (s *recordingCommentService) Create(_ context.Context, _ DBTX, in CreateInput) (Comment, error) {
	s.input = in
	if s.err != nil {
		return Comment{}, s.err
	}
	return Comment{ID: 9, PostID: in.PostID, UserID: in.UserID, Content: in.Content}, nil
}

func (s *recordingCommentService) List(_ context.Context, _ DBTX, postID int64) ([]Comment, error) {
	s.listPostID = postID
	return s.listed, nil
}
