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

type recordingCommentService struct {
	input CreateInput
}

func (s *recordingCommentService) Create(_ context.Context, _ DBTX, in CreateInput) (Comment, error) {
	s.input = in
	return Comment{ID: 9, PostID: in.PostID, UserID: in.UserID, Content: in.Content}, nil
}
