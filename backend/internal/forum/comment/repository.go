package comment

import (
	"context"

	"ai-forum/backend/internal/database"
)

type DBTX = database.DBTX

type Repository interface {
	Create(context.Context, DBTX, Comment) (Comment, error)
	IncrementCommentCount(context.Context, DBTX, int64) error
	SoftDelete(context.Context, DBTX, int64) error
	DecrementCommentCount(context.Context, DBTX, int64) error
}

type SQLRepository struct{}

func NewSQLRepository() *SQLRepository {
	return &SQLRepository{}
}

func (r *SQLRepository) Create(ctx context.Context, tx DBTX, c Comment) (Comment, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO comments (post_id, user_id, parent_comment_id, comment_type, ai_agent_id, content)
		VALUES (?, ?, ?, ?, ?, ?)`,
		c.PostID, c.UserID, c.ParentCommentID, c.CommentType, nil, c.Content)
	if err != nil {
		return Comment{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Comment{}, err
	}
	c.ID = id
	return c, nil
}

func (r *SQLRepository) IncrementCommentCount(ctx context.Context, tx DBTX, postID int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE posts SET comment_count = comment_count + 1 WHERE id = ?`, postID)
	return err
}

func (r *SQLRepository) SoftDelete(ctx context.Context, tx DBTX, commentID int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE comments SET deleted_at = NOW() WHERE id = ?`, commentID)
	return err
}

func (r *SQLRepository) DecrementCommentCount(ctx context.Context, tx DBTX, postID int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = ?`, postID)
	return err
}
