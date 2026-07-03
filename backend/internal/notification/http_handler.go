package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ai-forum/backend/internal/auth"
	"ai-forum/backend/internal/database"
)

type Notification struct {
	ID          int64           `db:"id" json:"id"`
	RecipientID int64           `db:"recipient_id" json:"recipient_id"`
	Type        string          `db:"type" json:"type"`
	Payload     json.RawMessage `db:"payload" json:"payload"`
	ReadAt      *time.Time      `db:"read_at" json:"read_at"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}

type HTTPHandler struct {
	db database.DBTX
}

func NewHTTPHandler(db database.DBTX) *HTTPHandler {
	return &HTTPHandler{db: db}
}

func (h *HTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := h.list(r.Context(), sub.UserID)
	if err != nil {
		http.Error(w, "list notifications", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(rows)
}

func (h *HTTPHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	count, err := h.unreadCount(r.Context(), sub.UserID)
	if err != nil {
		http.Error(w, "count notifications", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *HTTPHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("notificationId"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid notification id", http.StatusBadRequest)
		return
	}
	if err := h.markRead(r.Context(), sub.UserID, id); err != nil {
		http.Error(w, "mark notification read", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTPHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.markAllRead(r.Context(), sub.UserID); err != nil {
		http.Error(w, "mark notifications read", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTPHandler) list(ctx context.Context, userID int64) ([]Notification, error) {
	var rows []Notification
	err := h.db.SelectContext(ctx, &rows, `
		SELECT id, recipient_id, type, payload, read_at, created_at
		FROM notifications
		WHERE recipient_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 50`, userID)
	return rows, err
}

func (h *HTTPHandler) unreadCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := h.db.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM notifications
		WHERE recipient_id = ? AND read_at IS NULL`, userID)
	return count, err
}

func (h *HTTPHandler) markRead(ctx context.Context, userID, id int64) error {
	_, err := h.db.ExecContext(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, NOW())
		WHERE id = ? AND recipient_id = ?`, id, userID)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

func (h *HTTPHandler) markAllRead(ctx context.Context, userID int64) error {
	_, err := h.db.ExecContext(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, NOW())
		WHERE recipient_id = ? AND read_at IS NULL`, userID)
	return err
}
