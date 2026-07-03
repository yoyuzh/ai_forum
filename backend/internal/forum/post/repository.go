package post

import (
	"context"
	"database/sql"

	"ai-forum/backend/internal/database"
)

type DBTX = database.DBTX

type Repository interface {
	Create(ctx context.Context, tx DBTX, p Post) (Post, error)
	List(ctx context.Context, tx DBTX) ([]Post, error)
	Get(ctx context.Context, tx DBTX, postID int64) (Post, error)
	Update(ctx context.Context, tx DBTX, p Post) (Post, error)
	SoftDelete(ctx context.Context, tx DBTX, postID int64) error
	UpdateStatus(ctx context.Context, tx DBTX, postID int64, status string) error
}

type SnapshotDB interface {
	GetContext(context.Context, interface{}, string, ...interface{}) error
}

type SQLRepository struct{}

func NewSQLRepository() *SQLRepository {
	return &SQLRepository{}
}

func (r *SQLRepository) Create(ctx context.Context, tx DBTX, p Post) (Post, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO posts (author_id, title, content, status)
		VALUES (?, ?, ?, ?)`,
		p.AuthorID, p.Title, p.Content, p.Status)
	if err != nil {
		return Post{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Post{}, err
	}
	p.ID = id
	return p, nil
}

func (r *SQLRepository) List(ctx context.Context, tx DBTX) ([]Post, error) {
	var posts []Post
	err := tx.SelectContext(ctx, &posts, `
		SELECT id, author_id, title, content, status
		FROM posts
		WHERE deleted_at IS NULL
		ORDER BY id DESC`)
	return posts, err
}

func (r *SQLRepository) Get(ctx context.Context, tx DBTX, postID int64) (Post, error) {
	var p Post
	err := tx.GetContext(ctx, &p, `
		SELECT id, author_id, title, content, status
		FROM posts
		WHERE id = ? AND deleted_at IS NULL`, postID)
	return p, err
}

func (r *SQLRepository) Update(ctx context.Context, tx DBTX, p Post) (Post, error) {
	_, err := tx.ExecContext(ctx, `
		UPDATE posts
		SET title = ?, content = ?
		WHERE id = ? AND author_id = ? AND deleted_at IS NULL`, p.Title, p.Content, p.ID, p.AuthorID)
	if err != nil {
		return Post{}, err
	}
	return r.Get(ctx, tx, p.ID)
}

func (r *SQLRepository) SoftDelete(ctx context.Context, tx DBTX, postID int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL`, postID)
	return err
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, tx DBTX, postID int64, status string) error {
	_, err := tx.ExecContext(ctx, `UPDATE posts SET status = ? WHERE id = ? AND deleted_at IS NULL`, status, postID)
	return err
}

func LoadPostSnapshot(ctx context.Context, db SnapshotDB, postID int64) (PostSnapshot, error) {
	var p PostSnapshot
	err := db.GetContext(ctx, &p, `
		SELECT id, view_count, like_count, comment_count, ai_reply_count, created_at
		FROM posts
		WHERE id = ? AND deleted_at IS NULL`, postID)
	if err == sql.ErrNoRows {
		return PostSnapshot{}, err
	}
	return p, err
}
