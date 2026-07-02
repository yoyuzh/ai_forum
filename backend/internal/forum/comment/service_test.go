package comment

import (
	"context"
	"testing"

	"ai-forum/backend/internal/outbox"
)

func TestServiceCreateAndDeleteAppendOutbox(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{id: 9}
	var events []outbox.Event
	svc := NewService(repo, func(ctx context.Context, _ DBTX, event outbox.Event) error {
		events = append(events, event)
		return nil
	})

	c, err := svc.Create(context.Background(), tx, CreateInput{PostID: 42, UserID: 7, Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != 9 || !repo.incremented {
		t.Fatalf("comment=%#v incremented=%v", c, repo.incremented)
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
}

type recordingRepository struct {
	id          int64
	incremented bool
	decremented bool
	deleted     bool
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
