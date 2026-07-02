package favorite

import (
	"context"

	"ai-forum/backend/internal/database"
)

type DBTX = database.DBTX

type SQLRepository struct{}

func NewSQLRepository() *SQLRepository {
	return &SQLRepository{}
}

func (r *SQLRepository) Favorite(ctx context.Context, tx DBTX, userID, postID int64) (bool, error) {
	res, err := tx.ExecContext(ctx, `INSERT IGNORE INTO favorites (user_id, post_id) VALUES (?, ?)`, userID, postID)
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
	_, err = tx.ExecContext(ctx, `UPDATE posts SET favorite_count = (SELECT COUNT(*) FROM favorites WHERE post_id = ?) WHERE id = ?`, postID, postID)
	return true, err
}

func (r *SQLRepository) Unfavorite(ctx context.Context, tx DBTX, userID, postID int64) (bool, error) {
	res, err := tx.ExecContext(ctx, `DELETE FROM favorites WHERE user_id = ? AND post_id = ?`, userID, postID)
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
	_, err = tx.ExecContext(ctx, `UPDATE posts SET favorite_count = (SELECT COUNT(*) FROM favorites WHERE post_id = ?) WHERE id = ?`, postID, postID)
	return true, err
}

func (r *SQLRepository) List(ctx context.Context, tx DBTX, userID int64) ([]int64, error) {
	var ids []int64
	err := tx.SelectContext(ctx, &ids, `SELECT post_id FROM favorites WHERE user_id = ? ORDER BY id`, userID)
	return ids, err
}

func changedRows(res interface{ RowsAffected() (int64, error) }) (bool, error) {
	rows, err := res.RowsAffected()
	return rows > 0, err
}
