package post

import (
	"context"
	"database/sql"
	"fmt"

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
	if err := tx.SelectContext(ctx, &posts, `
		SELECT id, author_id, title, content, status, view_count, comment_count, like_count, ai_reply_count, created_at
		FROM posts
		WHERE deleted_at IS NULL
		ORDER BY id DESC`); err != nil {
		return nil, err
	}
	if err := r.loadTags(ctx, tx, posts); err != nil {
		return nil, err
	}
	if err := r.loadAIResponders(ctx, tx, posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *SQLRepository) Get(ctx context.Context, tx DBTX, postID int64) (Post, error) {
	var p Post
	err := tx.GetContext(ctx, &p, `
		SELECT id, author_id, title, content, status, view_count, comment_count, like_count, ai_reply_count, created_at
		FROM posts
		WHERE id = ? AND deleted_at IS NULL`, postID)
	if err != nil {
		return p, err
	}
	tags, err := r.tagsForPost(ctx, tx, postID)
	if err != nil {
		return Post{}, err
	}
	p.Tags = tags
	responders, err := r.aiRespondersForPost(ctx, tx, postID)
	if err != nil {
		return Post{}, err
	}
	p.AIResponders = responders
	return p, nil
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

func (r *SQLRepository) loadTags(ctx context.Context, tx DBTX, posts []Post) error {
	for i := range posts {
		tags, err := r.tagsForPost(ctx, tx, posts[i].ID)
		if err != nil {
			return fmt.Errorf("load post tags: %w", err)
		}
		posts[i].Tags = tags
	}
	return nil
}

func (r *SQLRepository) tagsForPost(ctx context.Context, tx DBTX, postID int64) ([]string, error) {
	var rows []struct {
		Name string `db:"tag_name"`
	}
	if err := tx.SelectContext(ctx, &rows, `
		SELECT tag_name
		FROM post_tags
		WHERE post_id = ?
		ORDER BY id`, postID); err != nil {
		return nil, err
	}
	tags := make([]string, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, row.Name)
	}
	return tags, nil
}

func (r *SQLRepository) loadAIResponders(ctx context.Context, tx DBTX, posts []Post) error {
	for i := range posts {
		responders, err := r.aiRespondersForPost(ctx, tx, posts[i].ID)
		if err != nil {
			return fmt.Errorf("load ai responders: %w", err)
		}
		posts[i].AIResponders = responders
	}
	return nil
}

func (r *SQLRepository) aiRespondersForPost(ctx context.Context, tx DBTX, postID int64) ([]AIResponder, error) {
	var rows []AIResponder
	if err := tx.SelectContext(ctx, &rows, `
		SELECT c.ai_agent_id, a.name
		FROM comments c
		JOIN ai_agents a ON a.id = c.ai_agent_id
		WHERE c.post_id = ?
		  AND c.comment_type = 'AI'
		  AND c.ai_agent_id IS NOT NULL
		  AND c.deleted_at IS NULL
		GROUP BY c.ai_agent_id, a.name
		ORDER BY MIN(c.id)
		LIMIT 3`, postID); err != nil {
		return nil, err
	}
	return rows, nil
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
