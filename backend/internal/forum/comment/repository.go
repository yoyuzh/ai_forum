package comment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/database"
)

type DBTX = database.DBTX

var ErrCommentNotFound = errors.New("comment not found")

type Repository interface {
	Create(context.Context, DBTX, Comment) (Comment, error)
	IncrementCommentCount(context.Context, DBTX, int64) error
	SoftDelete(context.Context, DBTX, int64) error
	DecrementCommentCount(context.Context, DBTX, int64) error
	FindMentionAgents(context.Context, DBTX, []string) ([]MentionAgent, error)
	CreateMention(context.Context, DBTX, CommentMention) error
	Get(context.Context, DBTX, int64) (Comment, error)
	ListByPost(context.Context, DBTX, int64) ([]Comment, error)
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

func (r *SQLRepository) FindMentionAgents(ctx context.Context, tx DBTX, names []string) ([]MentionAgent, error) {
	if len(names) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(`SELECT id, name, enabled, allow_mention FROM ai_agents WHERE name IN (?)`, names)
	if err != nil {
		return nil, err
	}
	var agents []MentionAgent
	if err := tx.SelectContext(ctx, &agents, query, args...); err != nil {
		return nil, err
	}
	return agents, nil
}

func (r *SQLRepository) CreateMention(ctx context.Context, tx DBTX, mention CommentMention) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO comment_mentions (comment_id, mentioned_ai_agent_id)
		VALUES (?, ?)`,
		mention.CommentID, mention.AIAgentID)
	return err
}

func (r *SQLRepository) Get(ctx context.Context, tx DBTX, id int64) (Comment, error) {
	var c Comment
	err := tx.GetContext(ctx, &c, `
		SELECT id, post_id, COALESCE(user_id, 0) AS user_id, parent_comment_id, comment_type, ai_agent_id, content
		FROM comments
		WHERE id = ? AND deleted_at IS NULL`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Comment{}, ErrCommentNotFound
	}
	return c, err
}

func (r *SQLRepository) ListByPost(ctx context.Context, tx DBTX, postID int64) ([]Comment, error) {
	var rows []struct {
		Comment
		UserDisplayName sql.NullString `db:"user_display_name"`
		UserUsername    sql.NullString `db:"user_username"`
		AIAgentName     sql.NullString `db:"ai_agent_name"`
	}
	err := tx.SelectContext(ctx, &rows, `
		SELECT c.id, c.post_id, COALESCE(c.user_id, 0) AS user_id, c.parent_comment_id,
		       c.comment_type, c.ai_agent_id, COALESCE(c.trigger_type, '') AS trigger_type, c.content,
		       u.display_name AS user_display_name, u.username AS user_username, a.name AS ai_agent_name
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		LEFT JOIN ai_agents a ON a.id = c.ai_agent_id
		WHERE c.post_id = ? AND c.deleted_at IS NULL
		ORDER BY c.id ASC`, postID)
	if err != nil {
		return nil, err
	}
	comments := make([]Comment, 0, len(rows))
	for _, row := range rows {
		c := row.Comment
		switch c.CommentType {
		case "AI":
			if row.AIAgentName.Valid {
				c.Author = &Author{Username: row.AIAgentName.String, IsAI: true}
			}
		default:
			name := ""
			if row.UserDisplayName.Valid && row.UserDisplayName.String != "" {
				name = row.UserDisplayName.String
			} else if row.UserUsername.Valid {
				name = row.UserUsername.String
			}
			if name != "" {
				c.Author = &Author{Username: name, IsAI: false}
			}
		}
		comments = append(comments, c)
	}
	return comments, nil
}
