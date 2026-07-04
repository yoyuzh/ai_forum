package search

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueryHandlerFiltersESHitsThroughMySQL(t *testing.T) {
	handler := NewQueryHandler(&queryDB{}, queryStore{ids: []int64{1, 2, 3}})
	req := httptest.NewRequest(http.MethodGet, "/api/search/posts?q=visible", nil)
	rec := httptest.NewRecorder()

	handler.SearchPosts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "visible") || strings.Contains(body, "hidden") || strings.Contains(body, "deleted") {
		t.Fatalf("body = %s, want only MySQL-visible hit", body)
	}
}

func TestQueryHandlerReturnsUnavailableWhenESFails(t *testing.T) {
	handler := NewQueryHandler(&queryDB{}, queryStore{err: sql.ErrConnDone})
	req := httptest.NewRequest(http.MethodGet, "/api/search/posts?q=visible", nil)
	rec := httptest.NewRecorder()

	handler.SearchPosts(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

type queryStore struct {
	ids []int64
	err error
}

func (s queryStore) Search(context.Context, string, int) ([]int64, error) {
	return s.ids, s.err
}

type queryDB struct{}

func (queryDB) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (queryDB) GetContext(context.Context, interface{}, string, ...interface{}) error {
	return nil
}

func (queryDB) SelectContext(_ context.Context, dest interface{}, query string, args ...interface{}) error {
	if !strings.Contains(query, "status = 'NORMAL'") || !strings.Contains(query, "deleted_at IS NULL") {
		return sql.ErrNoRows
	}
	rows := []PostResult{
		{ID: 1, AuthorID: 7, Title: "visible", Content: "body", Status: "NORMAL"},
	}
	*(dest.(*[]PostResult)) = rows
	_ = args
	return nil
}
