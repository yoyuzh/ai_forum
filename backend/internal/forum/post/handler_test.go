package post

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-forum/backend/internal/auth"
)

func TestHandlerCreateUsesAuthenticatedSubject(t *testing.T) {
	svc := &recordingPostService{}
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error { return fn(nil) })
	req := httptest.NewRequest(http.MethodPost, "/api/posts", strings.NewReader(`{"title":"hello","content":"body"}`))
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7, Username: "alice"}))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if svc.input.AuthorID != 7 || svc.input.Title != "hello" {
		t.Fatalf("input = %#v", svc.input)
	}
}

func TestHandlerUpdateStatusUsesPathIDAndTransaction(t *testing.T) {
	svc := &recordingPostService{}
	var ranTx bool
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error {
		ranTx = true
		return fn(nil)
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/posts/42/status", strings.NewReader(`{"status":"HIDDEN"}`))
	req.SetPathValue("postId", "42")
	rec := httptest.NewRecorder()

	h.UpdateStatus(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if !ranTx {
		t.Fatal("expected handler to run status update in transaction")
	}
	if svc.statusPostID != 42 || svc.status != "HIDDEN" {
		t.Fatalf("status update = (%d,%q), want (42,HIDDEN)", svc.statusPostID, svc.status)
	}
}

func TestHandlerReadUpdateDeletePosts(t *testing.T) {
	svc := &recordingPostService{}
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error { return fn(nil) })

	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/posts", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "hello") {
		t.Fatalf("list status/body = %d/%s", rec.Code, rec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/posts/42", nil)
	req.SetPathValue("postId", "42")
	rec = httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusOK || svc.gotPostID != 42 {
		t.Fatalf("get status/postID = %d/%d", rec.Code, svc.gotPostID)
	}

	req = httptest.NewRequest(http.MethodPatch, "/api/posts/42", strings.NewReader(`{"title":"new","content":"body"}`))
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7, Username: "alice"}))
	rec = httptest.NewRecorder()
	h.UpdateOwn(rec, req)
	if rec.Code != http.StatusOK || svc.update.PostID != 42 || svc.update.AuthorID != 7 {
		t.Fatalf("update status/input = %d/%#v", rec.Code, svc.update)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/posts/42", nil)
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7, Username: "alice"}))
	rec = httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusNoContent || svc.deletedPostID != 42 {
		t.Fatalf("delete status/postID = %d/%d", rec.Code, svc.deletedPostID)
	}
}

func TestHandlerListEncodesEmptyPostsAsArray(t *testing.T) {
	svc := &recordingPostService{returnNilList: true}
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error { return fn(nil) })
	rec := httptest.NewRecorder()

	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/posts", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("body = %q, want []", rec.Body.String())
	}
}

type recordingPostService struct {
	input         CreateInput
	statusPostID  int64
	status        string
	gotPostID     int64
	update        UpdateInput
	deletedPostID int64
	list          []Post
	returnNilList bool
}

func (s *recordingPostService) CreatePost(_ context.Context, _ DBTX, in CreateInput) (Post, error) {
	s.input = in
	return Post{ID: 42, AuthorID: in.AuthorID, Title: in.Title, Content: in.Content}, nil
}

func (s *recordingPostService) UpdateStatus(_ context.Context, _ DBTX, postID int64, status string) error {
	s.statusPostID = postID
	s.status = status
	return nil
}

func (s *recordingPostService) List(context.Context, DBTX) ([]Post, error) {
	if s.returnNilList {
		return nil, nil
	}
	if s.list != nil {
		return s.list, nil
	}
	return []Post{{ID: 42, AuthorID: 7, Title: "hello", Content: "body", Status: "NORMAL"}}, nil
}

func (s *recordingPostService) Get(_ context.Context, _ DBTX, postID int64) (Post, error) {
	s.gotPostID = postID
	return Post{ID: postID, AuthorID: 7, Title: "hello", Content: "body", Status: "NORMAL"}, nil
}

func (s *recordingPostService) UpdateOwn(_ context.Context, _ DBTX, in UpdateInput) (Post, error) {
	s.update = in
	return Post{ID: in.PostID, AuthorID: in.AuthorID, Title: in.Title, Content: in.Content, Status: "NORMAL"}, nil
}

func (s *recordingPostService) Delete(_ context.Context, _ DBTX, postID int64) error {
	s.deletedPostID = postID
	return nil
}
