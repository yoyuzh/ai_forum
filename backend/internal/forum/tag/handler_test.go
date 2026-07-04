package tag

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerListHotReturnsTopTags(t *testing.T) {
	service := &recordingHotTagLister{
		tags: []HotTag{{Name: "AI", PostCount: 4}, {Name: "Go", PostCount: 2}},
	}
	handler := NewHandler(service, runHotTagTx)
	rec := httptest.NewRecorder()

	handler.ListHot(rec, httptest.NewRequest(http.MethodGet, "/api/tags/hot?limit=2", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if service.limit != 2 {
		t.Fatalf("limit = %d, want 2", service.limit)
	}
	var got []HotTag
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "AI" || got[0].PostCount != 4 {
		t.Fatalf("tags = %#v", got)
	}
}

func TestHandlerListHotRejectsInvalidLimit(t *testing.T) {
	handler := NewHandler(&recordingHotTagLister{}, runHotTagTx)
	rec := httptest.NewRecorder()

	handler.ListHot(rec, httptest.NewRequest(http.MethodGet, "/api/tags/hot?limit=0", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerListHotReturnsServerError(t *testing.T) {
	handler := NewHandler(&recordingHotTagLister{err: errors.New("boom")}, runHotTagTx)
	rec := httptest.NewRecorder()

	handler.ListHot(rec, httptest.NewRequest(http.MethodGet, "/api/tags/hot", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

type recordingHotTagLister struct {
	limit int
	tags  []HotTag
	err   error
}

func (l *recordingHotTagLister) ListHotTags(_ context.Context, _ DBTX, limit int) ([]HotTag, error) {
	l.limit = limit
	return l.tags, l.err
}

func runHotTagTx(_ context.Context, fn func(DBTX) error) error {
	return fn(nil)
}
