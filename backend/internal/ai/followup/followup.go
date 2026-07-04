// Package followup judges whether an AI should answer a user reply.
package followup

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

type Candidate struct {
	AIAgentID int64  `db:"ai_agent_id"`
	Name      string `db:"name"`
	Content   string `db:"content"`
}

type Prompt struct {
	Post          Post
	ParentComment Comment
	ReplyComment  Comment
	Candidates    []Candidate
}

type Repository interface {
	LoadPost(context.Context, int64) (Post, error)
	LoadComment(context.Context, int64) (Comment, error)
	ListPostAICandidates(context.Context, int64) ([]Candidate, error)
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
	reply, err := h.repo.LoadComment(ctx, payload.ReplyCommentID)
	if err != nil {
		return err
	}
	if reply.CommentType != "USER" || reply.UserID == 0 {
		return nil
	}
	if payload.ParentCommentID == 0 {
		return h.handlePostComment(ctx, post, reply)
	}
	parent, err := h.repo.LoadComment(ctx, payload.ParentCommentID)
	if err != nil {
		return err
	}
	if parent.CommentType != "AI" || parent.AIAgentID == 0 {
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

func (h *Handler) handlePostComment(ctx context.Context, post Post, reply Comment) error {
	candidates, err := h.repo.ListPostAICandidates(ctx, post.ID)
	if err != nil {
		return err
	}
	allowed := map[int64]bool{}
	filtered := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.AIAgentID == 0 || allowed[candidate.AIAgentID] {
			continue
		}
		count, err := h.repo.CountFollowups(ctx, post.ID, candidate.AIAgentID)
		if err != nil {
			return err
		}
		if count >= 3 {
			continue
		}
		allowed[candidate.AIAgentID] = true
		filtered = append(filtered, candidate)
	}
	if len(filtered) == 0 {
		return nil
	}
	out, err := h.model.Generate(ctx, Prompt{Post: post, ReplyComment: reply, Candidates: filtered})
	if err != nil {
		return nil
	}
	ids, ok := parseAgentIDs(out)
	if !ok || len(ids) == 0 {
		return nil
	}
	replyID := reply.ID
	for _, id := range ids {
		if !allowed[id] {
			continue
		}
		if err := h.enqueuer.EnqueueGenerateAIReply(ctx, task.GenerateAIReplyPayload{
			PostID:          post.ID,
			ParentCommentID: &replyID,
			AIAgentID:       id,
			TriggerType:     "FOLLOWUP",
		}); err != nil {
			return err
		}
	}
	return nil
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

func parseAgentIDs(raw string) ([]int64, bool) {
	var obj struct {
		AgentIDs []int64 `json:"agent_ids"`
		AgentID  *int64  `json:"agent_id"`
	}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return nil, false
	}
	if obj.AgentID == nil && strings.Contains(raw, `"agent_id"`) {
		return nil, true
	}
	if obj.AgentID != nil {
		return []int64{*obj.AgentID}, true
	}
	if obj.AgentIDs == nil {
		return nil, false
	}
	return obj.AgentIDs, true
}

type SQLRepository struct {
	db sqlxDB
}

type sqlxDB interface {
	GetContext(context.Context, interface{}, string, ...interface{}) error
	SelectContext(context.Context, interface{}, string, ...interface{}) error
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

func (r *SQLRepository) ListPostAICandidates(ctx context.Context, postID int64) ([]Candidate, error) {
	var rows []Candidate
	err := r.db.SelectContext(ctx, &rows, `
		SELECT c.ai_agent_id, a.name, c.content
		FROM comments c
		JOIN ai_agents a ON a.id = c.ai_agent_id
		WHERE c.post_id = ? AND c.comment_type = 'AI' AND c.ai_agent_id IS NOT NULL
		  AND c.deleted_at IS NULL AND a.enabled = TRUE AND a.allow_followup = TRUE
		ORDER BY c.id DESC`, postID)
	if err != nil {
		return nil, err
	}
	seen := map[int64]bool{}
	out := make([]Candidate, 0, len(rows))
	for _, row := range rows {
		if seen[row.AIAgentID] {
			continue
		}
		seen[row.AIAgentID] = true
		out = append(out, row)
	}
	return out, nil
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
	if len(prompt.Candidates) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "Decide whether any existing AI participant should follow up to the user's new comment. Return JSON only: {\"agent_ids\":[candidate_id]} when one or more candidates should reply, or {\"agent_ids\":[]} / {\"agent_id\":null} when none should reply. Only choose ids listed in Candidates.\nPost: %s\nUser comment: %s\nCandidates:\n", prompt.Post.Title, prompt.ReplyComment.Content)
		for _, candidate := range prompt.Candidates {
			fmt.Fprintf(&b, "- id=%d name=%s previous_reply=%s\n", candidate.AIAgentID, candidate.Name, candidate.Content)
		}
		return m.client.Generate(ctx, modelclient.Request{
			Prompt:      b.String(),
			TaskID:      fmt.Sprint(prompt.ReplyComment.ID),
			TaskType:    "judge_ai_followup",
			PostID:      prompt.Post.ID,
			TriggerType: "FOLLOWUP",
		})
	}
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
