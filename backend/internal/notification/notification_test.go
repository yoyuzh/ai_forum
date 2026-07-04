package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestHandleAIReplyCompletedNotifiesPostAuthorAndDedups(t *testing.T) {
	db := &notificationDBTX{
		postAuthors: map[int64]int64{42: 7},
		processed:   map[string]bool{},
	}
	h := NewHandler(db)
	payload := mustJSON(t, EventPayload{EventID: "evt-ai-1", EventType: "ai.reply.completed", PostID: 42})

	if err := h.HandleSendNotification(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if err := h.HandleSendNotification(context.Background(), payload); err != nil {
		t.Fatal(err)
	}

	if len(db.notifications) != 1 {
		t.Fatalf("notifications = %d, want 1", len(db.notifications))
	}
	got := db.notifications[0]
	if got.recipientID != 7 || got.typ != "ai.reply.completed" || !strings.Contains(got.payload, `"post_id":42`) {
		t.Fatalf("notification = %#v, want post author ai reply notification", got)
	}
}

func TestHandleUserMentionedUsesPayloadRecipient(t *testing.T) {
	db := &notificationDBTX{processed: map[string]bool{}}
	h := NewHandler(db)
	payload := mustJSON(t, EventPayload{EventID: "evt-mentioned-1", EventType: "user.mentioned", PostID: 42, CommentID: 99, MentionedUserID: 8})

	if err := h.HandleSendNotification(context.Background(), payload); err != nil {
		t.Fatal(err)
	}

	if len(db.notifications) != 1 || db.notifications[0].recipientID != 8 {
		t.Fatalf("notifications = %#v, want mentioned user 8", db.notifications)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	body, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

type notificationDBTX struct {
	postAuthors   map[int64]int64
	processed     map[string]bool
	notifications []notificationRow
}

type notificationRow struct {
	recipientID int64
	typ         string
	payload     string
}

func (d *notificationDBTX) ExecContext(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case strings.Contains(query, "INSERT INTO processed_events"):
		d.processed[args[0].(string)+"/"+args[1].(string)] = true
	case strings.Contains(query, "INSERT INTO notifications"):
		d.notifications = append(d.notifications, notificationRow{
			recipientID: args[0].(int64),
			typ:         args[1].(string),
			payload:     string(args[2].([]byte)),
		})
	default:
		return nil, errors.New("unexpected exec: " + query)
	}
	return fakeNotificationResult(1), nil
}

func (d *notificationDBTX) GetContext(_ context.Context, dest interface{}, query string, args ...interface{}) error {
	switch {
	case strings.Contains(query, "processed_events"):
		key := args[0].(string) + "/" + args[1].(string)
		if d.processed[key] {
			*(dest.(*int)) = 1
		} else {
			*(dest.(*int)) = 0
		}
	case strings.Contains(query, "SELECT author_id FROM posts"):
		authorID, ok := d.postAuthors[args[0].(int64)]
		if !ok {
			return sql.ErrNoRows
		}
		*(dest.(*int64)) = authorID
	default:
		return errors.New("unexpected get: " + query)
	}
	return nil
}

func (d *notificationDBTX) SelectContext(context.Context, interface{}, string, ...interface{}) error {
	return errors.New("unexpected select")
}

type fakeNotificationResult int64

func (r fakeNotificationResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeNotificationResult) RowsAffected() (int64, error) { return int64(r), nil }
