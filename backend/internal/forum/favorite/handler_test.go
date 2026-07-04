package favorite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-forum/backend/internal/auth"
)

func TestHandlerFavoriteUnfavoriteUseAuthenticatedSubjectAndPostID(t *testing.T) {
	svc := &recordingFavoriteService{}
	txRuns := 0
	h := NewHandler(svc, func(ctx context.Context, fn func(DBTX) error) error {
		txRuns++
		return fn(nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/posts/42/favorite", nil)
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7}))
	rec := httptest.NewRecorder()

	h.Favorite(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("favorite status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if svc.favoritedUserID != 7 || svc.favoritedPostID != 42 || txRuns != 1 {
		t.Fatalf("favorite user=%d post=%d txRuns=%d, want 7/42/1", svc.favoritedUserID, svc.favoritedPostID, txRuns)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/posts/42/favorite", nil)
	req.SetPathValue("postId", "42")
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 7}))
	rec = httptest.NewRecorder()

	h.Unfavorite(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unfavorite status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if svc.unfavoritedUserID != 7 || svc.unfavoritedPostID != 42 || txRuns != 2 {
		t.Fatalf("unfavorite user=%d post=%d txRuns=%d, want 7/42/2", svc.unfavoritedUserID, svc.unfavoritedPostID, txRuns)
	}
}

type recordingFavoriteService struct {
	favoritedUserID   int64
	favoritedPostID   int64
	unfavoritedUserID int64
	unfavoritedPostID int64
}

func (s *recordingFavoriteService) Favorite(_ context.Context, _ DBTX, userID, postID int64) error {
	s.favoritedUserID = userID
	s.favoritedPostID = postID
	return nil
}

func (s *recordingFavoriteService) Unfavorite(_ context.Context, _ DBTX, userID, postID int64) error {
	s.unfavoritedUserID = userID
	s.unfavoritedPostID = postID
	return nil
}
