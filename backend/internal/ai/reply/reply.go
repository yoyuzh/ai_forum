// Package reply executes generate_ai_reply tasks.
package reply

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/ai/modelclient"
	"ai-forum/backend/internal/cache"
	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/event"
	"ai-forum/backend/internal/moderation"
	"ai-forum/backend/internal/outbox"
	"ai-forum/backend/internal/sse"
)

var ErrRateLimited = errors.New("ai reply rate limited")

type Task struct {
	PostID          int64
	ParentCommentID *int64
	AgentID         int64
	TriggerType     string
}

type Limiter interface {
	Allow(context.Context) (bool, error)
}

type HotCounter = cache.HotCounter

const HotCounterAIReply = cache.HotCounterAIReply

type HotTracker interface {
	RecordInteraction(context.Context, int64, HotCounter, int64) error
}

type Handler struct {
	db        *sqlx.DB
	model     modelclient.Client
	moderator moderation.Reviewer
	limiter   Limiter
	notifier  Notifier
	hot       HotTracker
}

type Event = sse.Event

type Notifier interface {
	Notify(context.Context, int64, Event) error
}

func NewSQLHandler(db *sqlx.DB, model modelclient.Client, moderator moderation.Reviewer, limiter Limiter) *Handler {
	return &Handler{db: db, model: model, moderator: moderator, limiter: limiter}
}

func (h *Handler) SetNotifier(notifier Notifier) {
	h.notifier = notifier
}

func (h *Handler) SetHotTracker(hot HotTracker) {
	h.hot = hot
}

func (h *Handler) HandleGenerateAIReply(ctx context.Context, task Task) error {
	if task.TriggerType == "" {
		task.TriggerType = "AUTO"
	}
	existing, err := h.existingStatus(ctx, task)
	if err != nil {
		return err
	}
	if existing != "" {
		return nil
	}
	if ok, err := h.limiter.Allow(ctx); err != nil {
		return err
	} else if !ok {
		return ErrRateLimited
	}
	taskID, inserted, err := h.createTask(ctx, task)
	if err != nil {
		return err
	}
	if !inserted {
		return nil
	}
	eligible, err := h.eligible(ctx, task)
	if err != nil {
		return err
	}
	if !eligible {
		return h.markTask(ctx, h.db, taskID, "SKIPPED", nil, nil)
	}
	if err := h.markTask(ctx, h.db, taskID, "RUNNING", nil, nil); err != nil {
		return err
	}
	promptInput, err := h.promptInput(ctx, task)
	if err != nil {
		return err
	}
	content, err := h.model.Generate(ctx, modelclient.Request{
		Prompt:      modelclient.BuildPrompt(promptInput),
		TaskID:      fmt.Sprint(taskID),
		TaskType:    "generate_ai_reply",
		PostID:      task.PostID,
		AIAgentID:   task.AgentID,
		TriggerType: task.TriggerType,
	})
	if err != nil {
		_ = h.persistFailure(ctx, taskID, task, err.Error())
		return err
	}
	content = strings.TrimSpace(content)
	result, err := h.moderator.Review(ctx, moderation.Input{Text: content})
	if err != nil {
		return err
	}
	if !result.Allowed {
		reason := result.Reason
		return h.markTask(ctx, h.db, taskID, "BLOCKED", &reason, nil)
	}
	if err := h.persistSuccess(ctx, taskID, task, content); err != nil {
		return err
	}
	if h.notifier != nil {
		return h.notifier.Notify(ctx, task.PostID, Event{Type: "ai_reply_completed"})
	}
	return nil
}

func (h *Handler) existingStatus(ctx context.Context, task Task) (string, error) {
	var status string
	err := h.db.GetContext(ctx, &status, `
		SELECT status
		FROM ai_reply_tasks
		WHERE post_id = ? AND parent_comment_id_norm = COALESCE(?,0)
		  AND ai_agent_id = ? AND trigger_type = ?
		LIMIT 1`,
		task.PostID, task.ParentCommentID, task.AgentID, task.TriggerType)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return status, err
}

func (h *Handler) createTask(ctx context.Context, task Task) (int64, bool, error) {
	res, err := h.db.ExecContext(ctx, `
		INSERT INTO ai_reply_tasks (post_id, parent_comment_id, ai_agent_id, trigger_type, status)
		VALUES (?, ?, ?, ?, 'PENDING')`,
		task.PostID, task.ParentCommentID, task.AgentID, task.TriggerType)
	if isDuplicateKey(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	id, err := res.LastInsertId()
	return id, true, err
}

func (h *Handler) promptInput(ctx context.Context, task Task) (modelclient.PromptInput, error) {
	var in modelclient.PromptInput
	if err := h.db.GetContext(ctx, &in, `
		SELECT a.name AS agent_name, p.title AS post_title, p.content AS post_content
		FROM posts p
		JOIN ai_agents a ON a.id = ?
		WHERE p.id = ? AND a.enabled = TRUE`,
		task.AgentID, task.PostID); err != nil {
		return in, fmt.Errorf("load prompt input: %w", err)
	}
	if task.ParentCommentID != nil {
		_ = h.db.GetContext(ctx, &in.ParentContent, `SELECT content FROM comments WHERE id = ?`, *task.ParentCommentID)
	}
	return in, nil
}

func (h *Handler) eligible(ctx context.Context, task Task) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM ai_agents WHERE id = ? AND enabled = TRUE`
	switch task.TriggerType {
	case "AUTO":
		query += ` AND allow_auto_reply = TRUE`
	case "MENTION":
		query += ` AND allow_mention = TRUE`
	case "FOLLOWUP":
		query += ` AND allow_followup = TRUE`
	}
	if err := h.db.GetContext(ctx, &count, query, task.AgentID); err != nil {
		return false, err
	}
	return count == 1, nil
}

func (h *Handler) persistSuccess(ctx context.Context, taskID int64, task Task, content string) error {
	return database.RunInTx(ctx, h.db, func(tx *sqlx.Tx) error {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO comments (post_id, user_id, parent_comment_id, comment_type, ai_agent_id, trigger_type, content)
			VALUES (?, NULL, ?, 'AI', ?, ?, ?)`,
			task.PostID, task.ParentCommentID, task.AgentID, task.TriggerType, content)
		if err != nil {
			return err
		}
		commentID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if err := outbox.Append(ctx, tx, outbox.Event{
			EventType:     event.AIReplyCompleted,
			AggregateType: "post",
			AggregateID:   task.PostID,
			Payload: map[string]any{
				"post_id":      task.PostID,
				"comment_id":   commentID,
				"ai_agent_id":  task.AgentID,
				"trigger_type": task.TriggerType,
			},
		}); err != nil {
			return err
		}
		if err := h.markTask(ctx, tx, taskID, "SUCCESS", nil, &commentID); err != nil {
			return err
		}
		if h.hot != nil {
			return h.hot.RecordInteraction(ctx, task.PostID, HotCounterAIReply, 1)
		}
		return nil
	})
}

func (h *Handler) persistFailure(ctx context.Context, taskID int64, task Task, reason string) error {
	return database.RunInTx(ctx, h.db, func(tx *sqlx.Tx) error {
		if err := h.markTask(ctx, tx, taskID, "FAILED", &reason, nil); err != nil {
			return err
		}
		return outbox.Append(ctx, tx, outbox.Event{
			EventType:     event.AIReplyFailed,
			AggregateType: "post",
			AggregateID:   task.PostID,
			Payload: map[string]any{
				"post_id":      task.PostID,
				"ai_agent_id":  task.AgentID,
				"trigger_type": task.TriggerType,
				"reason":       reason,
			},
		})
	})
}

func (h *Handler) markTask(ctx context.Context, db database.DBTX, taskID int64, status string, lastErr *string, commentID *int64) error {
	_, err := db.ExecContext(ctx, `
		UPDATE ai_reply_tasks
		SET status = ?, last_error = ?, comment_id = COALESCE(?, comment_id), attempt_count = attempt_count + 1
		WHERE id = ?`,
		status, lastErr, commentID, taskID)
	return err
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
