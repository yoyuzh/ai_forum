package like

import (
	"context"

	"ai-forum/backend/internal/database"
)

type DBTX = database.DBTX

type SQLRepository struct{}

func NewSQLRepository() *SQLRepository {
	return &SQLRepository{}
}

func (r *SQLRepository) Like(ctx context.Context, tx DBTX, userID, postID int64) (bool, error) {
	res, err := tx.ExecContext(ctx, `INSERT IGNORE INTO likes (user_id, post_id) VALUES (?, ?)`, userID, postID)
	if err != nil {
		return false, err
	}
	changed, err := changedRows(res)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	_, err = tx.ExecContext(ctx, `UPDATE posts SET like_count = (SELECT COUNT(*) FROM likes WHERE post_id = ?) WHERE id = ?`, postID, postID)
	return true, err
}

func (r *SQLRepository) Unlike(ctx context.Context, tx DBTX, userID, postID int64) (bool, error) {
	res, err := tx.ExecContext(ctx, `DELETE FROM likes WHERE user_id = ? AND post_id = ?`, userID, postID)
	if err != nil {
		return false, err
	}
	changed, err := changedRows(res)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	_, err = tx.ExecContext(ctx, `UPDATE posts SET like_count = (SELECT COUNT(*) FROM likes WHERE post_id = ?) WHERE id = ?`, postID, postID)
	return true, err
}

func (r *SQLRepository) Count(ctx context.Context, tx DBTX, postID int64) (int, error) {
	var count int
	err := tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM likes WHERE post_id = ?`, postID)
	return count, err
}

func changedRows(res interface{ RowsAffected() (int64, error) }) (bool, error) {
	rows, err := res.RowsAffected()
	return rows > 0, err
}
