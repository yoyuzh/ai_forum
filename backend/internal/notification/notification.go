// Package notification generates durable notification rows from domain events.
package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/event"
)

const consumerName = "asynq.send_notification"

type EventPayload struct {
	EventID         string `json:"event_id"`
	EventType       string `json:"event_type"`
	PostID          int64  `json:"post_id"`
	CommentID       int64  `json:"comment_id,omitempty"`
	MentionedUserID int64  `json:"mentioned_user_id,omitempty"`
}

type Handler struct {
	db database.DBTX
}

func NewHandler(db database.DBTX) *Handler {
	return &Handler{db: db}
}

func (h *Handler) HandleSendNotification(ctx context.Context, body []byte) error {
	var payload EventPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decode send_notification payload: %w", err)
	}
	if payload.EventID != "" {
		done, err := event.IsProcessed(ctx, h.db, payload.EventID, consumerName)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
	recipients, err := h.recipients(ctx, payload)
	if err != nil {
		return err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for _, recipientID := range recipients {
		if _, err := h.db.ExecContext(ctx, `
			INSERT INTO notifications (recipient_id, type, payload)
			VALUES (?, ?, ?)`,
			recipientID, payload.EventType, payloadJSON); err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
	}
	if payload.EventID != "" {
		return event.MarkProcessed(ctx, h.db, payload.EventID, consumerName)
	}
	return nil
}

func (h *Handler) recipients(ctx context.Context, payload EventPayload) ([]int64, error) {
	switch payload.EventType {
	case "user.mentioned":
		if payload.MentionedUserID <= 0 {
			return nil, nil
		}
		return []int64{payload.MentionedUserID}, nil
	case "comment.created", "ai.reply.completed":
		var authorID int64
		err := h.db.GetContext(ctx, &authorID, `SELECT author_id FROM posts WHERE id = ? AND deleted_at IS NULL`, payload.PostID)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("load post author: %w", err)
		}
		return []int64{authorID}, nil
	default:
		return nil, nil
	}
}
