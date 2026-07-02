package like

import (
	"context"
	"testing"

	"ai-forum/backend/internal/outbox"
)

func TestServiceLikeUnlikeCountAndOutbox(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{changed: true}
	var events []outbox.Event
	svc := NewService(repo, func(ctx context.Context, _ DBTX, event outbox.Event) error {
		events = append(events, event)
		return nil
	})

	if err := svc.Like(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}
	if err := svc.Unlike(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}

	count, err := svc.Count(context.Background(), tx, 42)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 || !repo.liked || !repo.unliked {
		t.Fatalf("count=%d liked=%v unliked=%v", count, repo.liked, repo.unliked)
	}
	if len(events) != 2 || events[0].EventType != "post.liked" || events[1].EventType != "post.unliked" {
		t.Fatalf("events = %#v", events)
	}

	repo.changed = false
	if err := svc.Like(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("noop like appended outbox event: %#v", events)
	}
}

type recordingRepository struct {
	liked   bool
	unliked bool
	count   int
	changed bool
}

func (r *recordingRepository) Like(context.Context, DBTX, int64, int64) (bool, error) {
	r.liked = true
	if !r.changed {
		return false, nil
	}
	r.count++
	return true, nil
}

func (r *recordingRepository) Unlike(context.Context, DBTX, int64, int64) (bool, error) {
	r.unliked = true
	r.count--
	return true, nil
}

func (r *recordingRepository) Count(context.Context, DBTX, int64) (int, error) {
	return r.count, nil
}
