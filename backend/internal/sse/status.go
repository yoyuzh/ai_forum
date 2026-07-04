package sse

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SQLStatusStore struct {
	db *sqlx.DB
}

func NewSQLStatusStore(db *sqlx.DB) *SQLStatusStore {
	return &SQLStatusStore{db: db}
}

func (s *SQLStatusStore) AIStatus(ctx context.Context, postID int64) (Status, error) {
	var rows []struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	if err := s.db.SelectContext(ctx, &rows, `
		SELECT status, COUNT(*) AS count
		FROM ai_reply_tasks
		WHERE post_id = ?
		GROUP BY status`, postID); err != nil {
		return Status{}, err
	}
	var out Status
	for _, row := range rows {
		switch row.Status {
		case "SUCCESS":
			out.CompletedCount += row.Count
		case "PENDING", "RUNNING", "RETRYING":
			out.RunningCount += row.Count
		case "FAILED":
			out.FailedCount += row.Count
		}
	}
	if err := s.db.GetContext(ctx, &out.RetryableCount, `
		SELECT COUNT(*)
		FROM ai_reply_tasks
		WHERE post_id = ? AND status = 'FAILED' AND attempt_count < 3`, postID); err != nil {
		return Status{}, err
	}
	return out, nil
}
