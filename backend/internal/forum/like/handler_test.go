package like

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-forum/backend/internal/auth"
)

func TestHandlerLikeUnlikeUseAuthenticatedSubjectAndPostID(t *testing.T) {
	svc := &recordingLikeService{}
	txRuns := 0
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error {
		txRuns++
		return fn(nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/posts/42/like", nil)
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7}))
	rec := httptest.NewRecorder()

	h.Like(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("like status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if svc.likedUserID != 7 || svc.likedPostID != 42 || txRuns != 1 {
		t.Fatalf("like user=%d post=%d txRuns=%d, want 7/42/1", svc.likedUserID, svc.likedPostID, txRuns)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/posts/42/like", nil)
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7}))
	rec = httptest.NewRecorder()

	h.Unlike(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unlike status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if svc.unlikedUserID != 7 || svc.unlikedPostID != 42 || txRuns != 2 {
		t.Fatalf("unlike user=%d post=%d txRuns=%d, want 7/42/2", svc.unlikedUserID, svc.unlikedPostID, txRuns)
	}
}

type recordingLikeService struct {
	likedUserID   int64
	likedPostID   int64
	unlikedUserID int64
	unlikedPostID int64
}

func (s *recordingLikeService) Like(_ context.Context, _ DBTX, userID, postID int64) error {
	s.likedUserID = userID
	s.likedPostID = postID
	return nil
}

func (s *recordingLikeService) Unlike(_ context.Context, _ DBTX, userID, postID int64) error {
	s.unlikedUserID = userID
	s.unlikedPostID = postID
	return nil
}
