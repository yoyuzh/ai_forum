package tag

import (
	"context"

	"ai-forum/backend/internal/database"
)

type DBTX = database.DBTX

type SQLRepository struct{}

func NewSQLRepository() *SQLRepository {
	return &SQLRepository{}
}

func (r *SQLRepository) Replace(ctx context.Context, tx DBTX, postID int64, tags []Tag) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM post_tags WHERE post_id = ?`, postID); err != nil {
		return err
	}
	for _, t := range tags {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO post_tags (post_id, tag_type, tag_name)
			VALUES (?, ?, ?)`, postID, t.Type, t.Name); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQLRepository) List(ctx context.Context, tx DBTX, postID int64) ([]Tag, error) {
	var rows []struct {
		PostID int64  `db:"post_id"`
		Type   string `db:"tag_type"`
		Name   string `db:"tag_name"`
	}
	if err := tx.SelectContext(ctx, &rows, `SELECT post_id, tag_type, tag_name FROM post_tags WHERE post_id = ? ORDER BY id`, postID); err != nil {
		return nil, err
	}
	tags := make([]Tag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, Tag{PostID: row.PostID, Type: row.Type, Name: row.Name})
	}
	return tags, nil
}
