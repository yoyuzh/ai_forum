package post

import (
	"context"
	"testing"

	"ai-forum/backend/internal/outbox"
)

func TestServiceCreatePostWritesPostAndOutboxInTx(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{id: 42}
	var events []outbox.Event
	svc := NewService(repo, func(ctx context.Context, _ DBTX, event outbox.Event) error {
		events = append(events, event)
		return nil
	})

	p, err := svc.CreatePost(context.Background(), tx, CreateInput{AuthorID: 7, Title: "Hello", Content: "Body"})
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != 42 || p.AuthorID != 7 {
		t.Fatalf("post = %#v", p)
	}
	if !repo.created {
		t.Fatal("expected repository create")
	}
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].EventType != "post.created" || events[0].AggregateType != "post" || events[0].AggregateID != 42 {
		t.Fatalf("event = %#v", events[0])
	}
}

func TestServiceUpdateStatusAppendsPostModerated(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{id: 42}
	var events []outbox.Event
	svc := NewService(repo, func(ctx context.Context, _ DBTX, event outbox.Event) error {
		events = append(events, event)
		return nil
	})

	if err := svc.UpdateStatus(context.Background(), tx, 42, "HIDDEN"); err != nil {
		t.Fatal(err)
	}
	if repo.status != "HIDDEN" {
		t.Fatalf("status = %q", repo.status)
	}
	if len(events) != 1 || events[0].EventType != "post.moderated" || events[0].AggregateID != 42 {
		t.Fatalf("events = %#v", events)
	}
}

func TestServiceReadUpdateDeletePosts(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{id: 42}
	var events []outbox.Event
	svc := NewService(repo, func(ctx context.Context, _ DBTX, event outbox.Event) error {
		events = append(events, event)
		return nil
	})

	listed, err := svc.List(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != 42 {
		t.Fatalf("listed = %#v", listed)
	}
	got, err := svc.Get(context.Background(), tx, 42)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 42 {
		t.Fatalf("got = %#v", got)
	}

	if _, err := svc.UpdateOwn(context.Background(), tx, UpdateInput{PostID: 42, AuthorID: 7, Title: "new", Content: "body"}); err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].EventType != "post.updated" {
		t.Fatalf("events after update = %#v", events)
	}
	if err := svc.Delete(context.Background(), tx, 42); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[1].EventType != "post.deleted" {
		t.Fatalf("events after delete = %#v", events)
	}
}

type recordingRepository struct {
	id      int64
	created bool
	status  string
	updated bool
	deleted bool
}

func (r *recordingRepository) Create(_ context.Context, _ DBTX, p Post) (Post, error) {
	r.created = true
	p.ID = r.id
	return p, nil
}

func (r *recordingRepository) UpdateStatus(_ context.Context, _ DBTX, postID int64, status string) error {
	r.id = postID
	r.status = status
	return nil
}

func (r *recordingRepository) List(context.Context, DBTX) ([]Post, error) {
	return []Post{{ID: r.id, AuthorID: 7, Title: "hello", Content: "body", Status: "NORMAL"}}, nil
}

func (r *recordingRepository) Get(context.Context, DBTX, int64) (Post, error) {
	return Post{ID: r.id, AuthorID: 7, Title: "hello", Content: "body", Status: "NORMAL"}, nil
}

func (r *recordingRepository) Update(_ context.Context, _ DBTX, p Post) (Post, error) {
	r.updated = true
	return p, nil
}

func (r *recordingRepository) SoftDelete(context.Context, DBTX, int64) error {
	r.deleted = true
	return nil
}
