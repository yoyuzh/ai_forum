// Package followup judges whether an AI should answer a user reply.
package followup

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-forum/backend/internal/ai/modelclient"
	"ai-forum/backend/internal/task"
)

type Post struct {
	ID      int64  `db:"id"`
	Title   string `db:"title"`
	Content string `db:"content"`
}

type Comment struct {
	ID          int64  `db:"id"`
	PostID      int64  `db:"post_id"`
	UserID      int64  `db:"user_id"`
	CommentType string `db:"comment_type"`
	AIAgentID   int64  `db:"ai_agent_id"`
	Content     string `db:"content"`
}

type Prompt struct {
	Post          Post
	ParentComment Comment
	ReplyComment  Comment
}

type Repository interface {
	LoadPost(context.Context, int64) (Post, error)
	LoadComment(context.Context, int64) (Comment, error)
	CountFollowups(context.Context, int64, int64) (int, error)
}

type Model interface {
	Generate(context.Context, Prompt) (string, error)
}

type GenerateEnqueuer interface {
	EnqueueGenerateAIReply(context.Context, task.GenerateAIReplyPayload) error
}

type Handler struct {
	repo     Repository
	model    Model
	enqueuer GenerateEnqueuer
}

func NewHandler(repo Repository, model Model, enqueuer GenerateEnqueuer) *Handler {
	return &Handler{repo: repo, model: model, enqueuer: enqueuer}
}

func (h *Handler) HandleJudgeAIFollowup(ctx context.Context, payload task.JudgeAIFollowupPayload) error {
	post, err := h.repo.LoadPost(ctx, payload.PostID)
	if err != nil {
		return err
	}
	parent, err := h.repo.LoadComment(ctx, payload.ParentCommentID)
	if err != nil {
		return err
	}
	reply, err := h.repo.LoadComment(ctx, payload.ReplyCommentID)
	if err != nil {
		return err
	}
	if parent.CommentType != "AI" || parent.AIAgentID == 0 || reply.CommentType != "USER" || reply.UserID == 0 {
		return nil
	}
	count, err := h.repo.CountFollowups(ctx, post.ID, parent.AIAgentID)
	if err != nil {
		return err
	}
	if count >= 3 {
		return nil
	}
	out, err := h.model.Generate(ctx, Prompt{Post: post, ParentComment: parent, ReplyComment: reply})
	if err != nil {
		return nil
	}
	should, ok := parseShouldReply(out)
	if !ok || !should {
		return nil
	}
	replyID := reply.ID
	return h.enqueuer.EnqueueGenerateAIReply(ctx, task.GenerateAIReplyPayload{
		PostID:          post.ID,
		ParentCommentID: &replyID,
		AIAgentID:       parent.AIAgentID,
		TriggerType:     "FOLLOWUP",
	})
}

func parseShouldReply(raw string) (bool, bool) {
	var obj struct {
		ShouldReply *bool  `json:"should_reply"`
		Reason      string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return false, false
	}
	if obj.ShouldReply == nil {
		return false, false
	}
	if obj.Reason == "" {
		return false, false
	}
	return *obj.ShouldReply, true
}

type SQLRepository struct {
	db sqlxDB
}

type sqlxDB interface {
	GetContext(context.Context, interface{}, string, ...interface{}) error
}

func NewSQLRepository(db sqlxDB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) LoadPost(ctx context.Context, id int64) (Post, error) {
	var post Post
	err := r.db.GetContext(ctx, &post, `SELECT id, title, content FROM posts WHERE id = ?`, id)
	return post, err
}

func (r *SQLRepository) LoadComment(ctx context.Context, id int64) (Comment, error) {
	var c Comment
	err := r.db.GetContext(ctx, &c, `
		SELECT id, post_id, COALESCE(user_id, 0) AS user_id, comment_type, COALESCE(ai_agent_id, 0) AS ai_agent_id, content
		FROM comments
		WHERE id = ? AND deleted_at IS NULL`, id)
	return c, err
}

func (r *SQLRepository) CountFollowups(ctx context.Context, postID, agentID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM ai_reply_tasks
		WHERE post_id = ? AND ai_agent_id = ? AND trigger_type = 'FOLLOWUP'`,
		postID, agentID)
	return count, err
}

type ModelClient struct {
	client modelclient.Client
}

func NewModelClient(client modelclient.Client) *ModelClient {
	return &ModelClient{client: client}
}

func (m *ModelClient) Generate(ctx context.Context, prompt Prompt) (string, error) {
	return m.client.Generate(ctx, modelclient.Request{Prompt: fmt.Sprintf(
		"Return JSON {\"should_reply\":boolean,\"reason\":string}. Post: %s\nAI: %s\nUser: %s",
		prompt.Post.Title, prompt.ParentComment.Content, prompt.ReplyComment.Content,
	),
		TaskID:      fmt.Sprint(prompt.ReplyComment.ID),
		TaskType:    "judge_ai_followup",
		PostID:      prompt.Post.ID,
		AIAgentID:   prompt.ParentComment.AIAgentID,
		TriggerType: "FOLLOWUP",
	})
}
