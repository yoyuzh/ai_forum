package notification

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-forum/backend/internal/auth"
)

func TestHTTPHandlerListsUnreadAndMarksReadForCurrentUser(t *testing.T) {
	db := &httpNotificationDB{rows: []Notification{
		{ID: 1, RecipientID: 7, Type: "ai.reply.completed", Payload: []byte(`{"title":"AI replied"}`), CreatedAt: time.Unix(1, 0)},
		{ID: 2, RecipientID: 9, Type: "comment.created", Payload: []byte(`{"title":"other user"}`), CreatedAt: time.Unix(2, 0)},
	}}
	h := NewHTTPHandler(db)

	req := authedRequest(http.MethodGet, "/api/notifications", 7)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"id":1`) || strings.Contains(rec.Body.String(), `"id":2`) {
		t.Fatalf("list body = %q, want current user row only", rec.Body.String())
	}

	req = authedRequest(http.MethodGet, "/api/notifications/unread-count", 7)
	rec = httptest.NewRecorder()
	h.UnreadCount(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"count":1`) {
		t.Fatalf("unread response code=%d body=%q, want count 1", rec.Code, rec.Body.String())
	}

	req = authedRequest(http.MethodPut, "/api/notifications/1/read", 7)
	req.SetPathValue("notificationId", "1")
	rec = httptest.NewRecorder()
	h.MarkRead(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("mark status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	req = authedRequest(http.MethodGet, "/api/notifications/unread-count", 7)
	rec = httptest.NewRecorder()
	h.UnreadCount(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"count":0`) {
		t.Fatalf("unread after mark code=%d body=%q, want count 0", rec.Code, rec.Body.String())
	}
}

func TestHTTPHandlerMarkAllReadOnlyCurrentUser(t *testing.T) {
	db := &httpNotificationDB{rows: []Notification{
		{ID: 1, RecipientID: 7, Type: "ai.reply.completed", Payload: []byte(`{}`), CreatedAt: time.Unix(1, 0)},
		{ID: 2, RecipientID: 7, Type: "comment.created", Payload: []byte(`{}`), CreatedAt: time.Unix(2, 0)},
		{ID: 3, RecipientID: 9, Type: "comment.created", Payload: []byte(`{}`), CreatedAt: time.Unix(3, 0)},
	}}
	h := NewHTTPHandler(db)

	req := authedRequest(http.MethodPut, "/api/notifications/read-all", 7)
	rec := httptest.NewRecorder()
	h.MarkAllRead(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("mark all status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if db.rows[0].ReadAt == nil || db.rows[1].ReadAt == nil || db.rows[2].ReadAt != nil {
		t.Fatalf("readAt = %#v, want only user 7 rows marked", db.rows)
	}
}

func authedRequest(method, target string, userID int64) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	return req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: userID}))
}

type httpNotificationDB struct {
	rows []Notification
}

func (d *httpNotificationDB) SelectContext(_ context.Context, dest interface{}, query string, args ...interface{}) error {
	rows := dest.(*[]Notification)
	userID := args[0].(int64)
	for _, row := range d.rows {
		if row.RecipientID == userID {
			*rows = append(*rows, row)
		}
	}
	return nil
}

func (d *httpNotificationDB) GetContext(_ context.Context, dest interface{}, query string, args ...interface{}) error {
	count := dest.(*int)
	userID := args[0].(int64)
	for _, row := range d.rows {
		if row.RecipientID == userID && row.ReadAt == nil {
			(*count)++
		}
	}
	return nil
}

func (d *httpNotificationDB) ExecContext(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	now := time.Unix(99, 0)
	if strings.Contains(query, "WHERE id = ? AND recipient_id = ?") {
		notificationID := args[0].(int64)
		userID := args[1].(int64)
		for i := range d.rows {
			if d.rows[i].ID == notificationID && d.rows[i].RecipientID == userID {
				d.rows[i].ReadAt = &now
			}
		}
		return fakeResult(1), nil
	}
	userID := args[0].(int64)
	for i := range d.rows {
		if d.rows[i].RecipientID == userID {
			d.rows[i].ReadAt = &now
		}
	}
	return fakeResult(1), nil
}

type fakeResult int64

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return int64(r), nil }
