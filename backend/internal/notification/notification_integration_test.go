//go:build integration

package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/auth"
	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/database"
)

func TestAIReplyCompletedWritesOneNotificationOnRedelivery(t *testing.T) {
	db, m := notificationTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()
	res, err := db.ExecContext(ctx, `INSERT INTO posts (author_id, title, content, status) VALUES (1, 'notify', 'body', 'NORMAL')`)
	require.NoError(t, err)
	postID, err := res.LastInsertId()
	require.NoError(t, err)
	eventID := fmt.Sprintf("evt-p9-notify-live-%d", time.Now().UnixNano())
	body, err := json.Marshal(EventPayload{EventID: eventID, EventType: "ai.reply.completed", PostID: postID})
	require.NoError(t, err)
	handler := NewHandler(db)

	require.NoError(t, handler.HandleSendNotification(ctx, body))
	require.NoError(t, handler.HandleSendNotification(ctx, body))

	var rows int
	require.NoError(t, db.GetContext(ctx, &rows, `
		SELECT COUNT(*) FROM notifications
		WHERE recipient_id = 1 AND type = 'ai.reply.completed'
		  AND JSON_UNQUOTE(JSON_EXTRACT(payload, '$.event_id')) = ?`, eventID))
	require.Equal(t, 1, rows)
}

func TestHTTPHandlerMarksNotificationReadIntegration(t *testing.T) {
	db, m := notificationTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()
	res, err := db.ExecContext(ctx, `
		INSERT INTO notifications (recipient_id, type, payload)
		VALUES (1, 'ai.reply.completed', JSON_OBJECT('post_id', 1))`)
	require.NoError(t, err)
	notificationID, err := res.LastInsertId()
	require.NoError(t, err)

	h := NewHTTPHandler(db)
	req := httptest.NewRequest(http.MethodPut, "/api/notifications/1/read", nil)
	req.SetPathValue("notificationId", fmt.Sprint(notificationID))
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 1}))
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%q, want 204", rec.Code, rec.Body.String())
	}
	var unread int
	require.NoError(t, db.GetContext(ctx, &unread, `
		SELECT COUNT(*) FROM notifications WHERE id = ? AND read_at IS NULL`, notificationID))
	require.Equal(t, 0, unread)
}

func notificationTestDB(t *testing.T) (*sqlx.DB, *migrate.Migrate) {
	t.Helper()
	cfg := notificationMySQLCfgFromEnv()
	db, err := database.NewMySQL(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	wd, err := os.Getwd()
	require.NoError(t, err)
	src := "file://" + filepath.Join(wd, "..", "..", "migrations")
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	m, err := migrate.New(src, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })
	return db, m
}

func notificationMySQLCfgFromEnv() config.MySQLConfig {
	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}
	port := 3306
	if v := os.Getenv("MYSQL_PORT"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &port); err != nil {
			port = 3306
		}
	}
	return config.MySQLConfig{
		Host:     get("MYSQL_HOST", "127.0.0.1"),
		Port:     port,
		Username: get("MYSQL_USERNAME", "root"),
		Password: get("MYSQL_PASSWORD", "ai_forum_root"),
		Database: get("MYSQL_DATABASE", "ai_forum"),
	}
}
