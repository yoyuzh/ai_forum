package favorite

import (
	"context"
	"testing"

	"ai-forum/backend/internal/outbox"
)

func TestServiceFavoriteUnfavoriteListAndOutbox(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{changed: true}
	var events []outbox.Event
	svc := NewService(repo, func(ctx context.Context, _ DBTX, event outbox.Event) error {
		events = append(events, event)
		return nil
	})

	if err := svc.Favorite(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}
	if err := svc.Unfavorite(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}
	posts, err := svc.List(context.Background(), tx, 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0] != 42 || !repo.favorited || !repo.unfavorited {
		t.Fatalf("favorites = %#v", posts)
	}
	if len(events) != 2 || events[0].EventType != "post.favorited" || events[1].EventType != "post.unfavorited" {
		t.Fatalf("events = %#v", events)
	}

	repo.changed = false
	if err := svc.Favorite(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("noop favorite appended outbox event: %#v", events)
	}
}

type recordingRepository struct {
	favorited   bool
	unfavorited bool
	changed     bool
}

func (r *recordingRepository) Favorite(context.Context, DBTX, int64, int64) (bool, error) {
	r.favorited = true
	return r.changed, nil
}

func (r *recordingRepository) Unfavorite(context.Context, DBTX, int64, int64) (bool, error) {
	r.unfavorited = true
	return true, nil
}

func (r *recordingRepository) List(context.Context, DBTX, int64) ([]int64, error) {
	return []int64{42}, nil
}
