package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestHandlePostCreatedBuildsDocumentFromDatabase(t *testing.T) {
	store := &recordingIndexStore{}
	db := &searchDBTX{
		posts: map[int64]indexedPost{
			42: {ID: 42, AuthorID: 7, Title: "fresh title", Content: "fresh body", Status: "NORMAL"},
		},
		processed: map[string]bool{},
	}
	h := NewSyncHandler(db, store)
	payload := mustSearchJSON(t, SyncPayload{
		EventID:   "evt-post-1",
		EventType: "post.created",
		PostID:    42,
		Title:     "stale payload title",
	})

	if err := h.HandleSyncSearchIndex(context.Background(), payload); err != nil {
		t.Fatal(err)
	}

	got := store.upserts["post:42"]
	if got.Title != "fresh title" || got.Body != "fresh body" || got.AuthorID != 7 {
		t.Fatalf("document = %#v, want DB values", got)
	}
}

func TestHandlePostCreatedDedupsByProcessedEvents(t *testing.T) {
	store := &recordingIndexStore{}
	db := &searchDBTX{
		posts:     map[int64]indexedPost{42: {ID: 42, Title: "title", Content: "body"}},
		processed: map[string]bool{},
	}
	h := NewSyncHandler(db, store)
	payload := mustSearchJSON(t, SyncPayload{EventID: "evt-search-dedup-1", EventType: "post.created", PostID: 42})

	if err := h.HandleSyncSearchIndex(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if err := h.HandleSyncSearchIndex(context.Background(), payload); err != nil {
		t.Fatal(err)
	}

	if store.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d, want 1", store.upsertCalls)
	}
}

func TestHandleDeletedPostDeletesDocument(t *testing.T) {
	store := &recordingIndexStore{}
	h := NewSyncHandler(&searchDBTX{}, store)
	payload := mustSearchJSON(t, SyncPayload{EventID: "evt-post-delete-1", EventType: "post.deleted", PostID: 42})

	if err := h.HandleSyncSearchIndex(context.Background(), payload); err != nil {
		t.Fatal(err)
	}

	if len(store.deleted) != 1 || store.deleted[0] != "post:42" {
		t.Fatalf("deleted = %#v, want post:42", store.deleted)
	}
}

func TestBuildAllDocumentsMatchesIncrementalDocument(t *testing.T) {
	db := &searchDBTX{
		posts: map[int64]indexedPost{
			42: {ID: 42, AuthorID: 7, Title: "title", Content: "body", Status: "NORMAL"},
		},
		comments: map[int64]indexedComment{
			99: {ID: 99, PostID: 42, CommentType: "USER", Content: "comment"},
		},
	}
	h := NewSyncHandler(db, &recordingIndexStore{})

	incremental, err := h.BuildPostDocument(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	all, err := h.BuildAllDocuments(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	commentDoc, err := h.BuildCommentDocument(context.Background(), 99)
	if err != nil {
		t.Fatal(err)
	}

	if len(all) != 2 || all[0] != incremental || all[1] != commentDoc {
		t.Fatalf("rebuild docs = %#v, want post/comment incremental docs", all)
	}
}

func TestRebuildAllEnsuresIndexAndUpsertsAllDocuments(t *testing.T) {
	store := &recordingIndexStore{}
	db := &searchDBTX{
		posts: map[int64]indexedPost{
			42: {ID: 42, AuthorID: 7, Title: "title", Content: "body", Status: "NORMAL"},
		},
		comments: map[int64]indexedComment{
			99: {ID: 99, PostID: 42, CommentType: "USER", Content: "comment"},
		},
	}
	h := NewSyncHandler(db, store)

	if err := h.RebuildAll(context.Background()); err != nil {
		t.Fatal(err)
	}

	if store.ensureCalls != 1 {
		t.Fatalf("ensure calls = %d, want 1", store.ensureCalls)
	}
	if len(store.upserts) != 2 || store.upserts["post:42"].Title != "title" || store.upserts["comment:99"].Body != "comment" {
		t.Fatalf("upserts = %#v, want rebuilt post/comment docs", store.upserts)
	}
}

func mustSearchJSON(t *testing.T, v any) []byte {
	t.Helper()
	body, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

type recordingIndexStore struct {
	ensureCalls int
	upserts     map[string]Document
	upsertCalls int
	deleted     []string
}

func (s *recordingIndexStore) EnsureIndex(context.Context) error {
	s.ensureCalls++
	return nil
}

func (s *recordingIndexStore) Upsert(_ context.Context, doc Document) error {
	s.upsertCalls++
	if s.upserts == nil {
		s.upserts = map[string]Document{}
	}
	s.upserts[doc.ID] = doc
	return nil
}

func (s *recordingIndexStore) Delete(_ context.Context, id string) error {
	s.deleted = append(s.deleted, id)
	return nil
}

type searchDBTX struct {
	posts     map[int64]indexedPost
	comments  map[int64]indexedComment
	processed map[string]bool
}

func (d *searchDBTX) ExecContext(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	if strings.Contains(query, "INSERT INTO processed_events") {
		if d.processed == nil {
			d.processed = map[string]bool{}
		}
		d.processed[args[0].(string)+"/"+args[1].(string)] = true
		return fakeSearchResult(1), nil
	}
	return nil, errors.New("unexpected exec")
}

func (d *searchDBTX) GetContext(_ context.Context, dest interface{}, query string, args ...interface{}) error {
	if strings.Contains(query, "processed_events") {
		if d.processed[args[0].(string)+"/"+args[1].(string)] {
			*(dest.(*int)) = 1
		} else {
			*(dest.(*int)) = 0
		}
		return nil
	}
	if strings.Contains(query, "FROM comments") {
		comment, ok := d.comments[args[0].(int64)]
		if !ok {
			return sql.ErrNoRows
		}
		*(dest.(*indexedComment)) = comment
		return nil
	}
	if !strings.Contains(query, "FROM posts") {
		return errors.New("unexpected get: " + query)
	}
	post, ok := d.posts[args[0].(int64)]
	if !ok {
		return sql.ErrNoRows
	}
	*(dest.(*indexedPost)) = post
	return nil
}

func (d *searchDBTX) SelectContext(_ context.Context, dest interface{}, query string, _ ...interface{}) error {
	if strings.Contains(query, "FROM comments") {
		var comments []indexedComment
		for _, comment := range d.comments {
			comments = append(comments, comment)
		}
		*(dest.(*[]indexedComment)) = comments
		return nil
	}
	if !strings.Contains(query, "FROM posts") {
		return errors.New("unexpected select: " + query)
	}
	var posts []indexedPost
	for _, post := range d.posts {
		posts = append(posts, post)
	}
	*(dest.(*[]indexedPost)) = posts
	return nil
}

type fakeSearchResult int64

func (r fakeSearchResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeSearchResult) RowsAffected() (int64, error) { return int64(r), nil }
