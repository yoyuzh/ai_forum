package comment

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"ai-forum/backend/internal/cache"
	"ai-forum/backend/internal/outbox"
	"ai-forum/backend/internal/task"
)

var (
	ErrMentionRateLimited = errors.New("mention rate limited")
	ErrTooManyAIMentions  = errors.New("too many AI mentions")
)

type CreateInput struct {
	PostID          int64
	UserID          int64
	ParentCommentID *int64
	Content         string
}

type AppendFunc func(context.Context, DBTX, outbox.Event) error
type AfterCommitFunc func(func(context.Context) error)
type afterCommitContextKey struct{}
type HotCounter = cache.HotCounter

const HotCounterComment = cache.HotCounterComment

type MentionLimiter interface {
	AllowMentions(context.Context, int64, int) error
}

type GenerateEnqueuer interface {
	EnqueueGenerateAIReply(context.Context, task.GenerateAIReplyPayload) error
}

type FollowupEnqueuer interface {
	EnqueueJudgeAIFollowup(context.Context, task.JudgeAIFollowupPayload) error
}

type HotTracker interface {
	RecordInteraction(context.Context, int64, HotCounter, int64) error
}

type Service struct {
	repo      Repository
	append    AppendFunc
	mentionRL MentionLimiter
	after     AfterCommitFunc
	generate  GenerateEnqueuer
	followup  FollowupEnqueuer
	hot       HotTracker
}

type Option func(*Service)

func WithMentionLimiter(limiter MentionLimiter) Option {
	return func(s *Service) { s.mentionRL = limiter }
}

func WithAfterCommit(after AfterCommitFunc) Option {
	return func(s *Service) { s.after = after }
}

func WithGenerateEnqueuer(enqueuer GenerateEnqueuer) Option {
	return func(s *Service) { s.generate = enqueuer }
}

func WithFollowupEnqueuer(enqueuer FollowupEnqueuer) Option {
	return func(s *Service) { s.followup = enqueuer }
}

func WithHotTracker(hot HotTracker) Option {
	return func(s *Service) { s.hot = hot }
}

func NewService(repo Repository, append AppendFunc, opts ...Option) *Service {
	s := &Service{repo: repo, append: append}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) Create(ctx context.Context, tx DBTX, in CreateInput) (Comment, error) {
	content := strings.TrimSpace(in.Content)
	if in.PostID <= 0 || in.UserID <= 0 || content == "" {
		return Comment{}, fmt.Errorf("invalid comment")
	}
	agents, err := s.validMentionAgents(ctx, tx, content)
	if err != nil {
		return Comment{}, err
	}
	if len(agents) > 3 {
		return Comment{}, ErrTooManyAIMentions
	}
	if len(agents) > 0 && s.mentionRL != nil {
		if err := s.mentionRL.AllowMentions(ctx, in.UserID, len(agents)); err != nil {
			return Comment{}, err
		}
	}
	c, err := s.repo.Create(ctx, tx, Comment{
		PostID:          in.PostID,
		UserID:          in.UserID,
		ParentCommentID: in.ParentCommentID,
		CommentType:     "USER",
		Content:         content,
	})
	if err != nil {
		return Comment{}, err
	}
	if s.hot != nil {
		if err := s.hot.RecordInteraction(ctx, in.PostID, HotCounterComment, 1); err != nil {
			return Comment{}, err
		}
	}
	if err := s.append(ctx, tx, outbox.Event{
		EventType:     "comment.created",
		AggregateType: "comment",
		AggregateID:   c.ID,
		Payload:       map[string]any{"comment_id": c.ID, "post_id": c.PostID},
	}); err != nil {
		return Comment{}, err
	}
	for _, agent := range agents {
		if err := s.repo.CreateMention(ctx, tx, CommentMention{CommentID: c.ID, AIAgentID: agent.ID}); err != nil {
			return Comment{}, err
		}
		if s.generate != nil {
			agentID := agent.ID
			commentID := c.ID
			s.afterCommit(ctx, func(afterCtx context.Context) error {
				return s.generate.EnqueueGenerateAIReply(afterCtx, task.GenerateAIReplyPayload{
					PostID:          c.PostID,
					ParentCommentID: &commentID,
					AIAgentID:       agentID,
					TriggerType:     "MENTION",
				})
			})
		}
	}
	if len(agents) == 0 && s.followup != nil {
		var parentID int64
		if in.ParentCommentID != nil {
			parent, err := s.repo.Get(ctx, tx, *in.ParentCommentID)
			if err != nil && !errors.Is(err, ErrCommentNotFound) {
				return Comment{}, err
			}
			if parent.CommentType == "AI" {
				parentID = *in.ParentCommentID
			}
		}
		replyID := c.ID
		s.afterCommit(ctx, func(afterCtx context.Context) error {
			return s.followup.EnqueueJudgeAIFollowup(afterCtx, task.JudgeAIFollowupPayload{
				PostID:          c.PostID,
				ParentCommentID: parentID,
				ReplyCommentID:  replyID,
			})
		})
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, tx DBTX, postID int64) ([]Comment, error) {
	if postID <= 0 {
		return nil, fmt.Errorf("invalid post id")
	}
	return s.repo.ListByPost(ctx, tx, postID)
}

func (s *Service) Delete(ctx context.Context, tx DBTX, postID, commentID int64) error {
	if postID <= 0 || commentID <= 0 {
		return fmt.Errorf("invalid comment delete")
	}
	if err := s.repo.SoftDelete(ctx, tx, commentID); err != nil {
		return err
	}
	if err := s.repo.DecrementCommentCount(ctx, tx, postID); err != nil {
		return err
	}
	return s.append(ctx, tx, outbox.Event{
		EventType:     "comment.deleted",
		AggregateType: "comment",
		AggregateID:   commentID,
		Payload:       map[string]any{"comment_id": commentID, "post_id": postID},
	})
}

var mentionRE = regexp.MustCompile(`@([\pL\pN_]+)`)

func (s *Service) validMentionAgents(ctx context.Context, tx DBTX, content string) ([]MentionAgent, error) {
	names := parseMentionNames(content)
	if len(names) == 0 {
		return nil, nil
	}
	agents, err := s.repo.FindMentionAgents(ctx, tx, names)
	if err != nil {
		return nil, err
	}
	var out []MentionAgent
	for _, agent := range agents {
		if agent.Enabled && agent.AllowMention {
			out = append(out, agent)
		}
	}
	return out, nil
}

func parseMentionNames(content string) []string {
	matches := mentionRE.FindAllStringSubmatch(content, -1)
	seen := map[string]bool{}
	var names []string
	for _, match := range matches {
		name := match[1]
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

func (s *Service) afterCommit(ctx context.Context, fn func(context.Context) error) {
	if s.after != nil {
		s.after(fn)
		return
	}
	addAfterCommit(ctx, fn)
}

func contextWithAfterCommit(ctx context.Context, add func(func(context.Context) error)) context.Context {
	return context.WithValue(ctx, afterCommitContextKey{}, add)
}

func addAfterCommit(ctx context.Context, fn func(context.Context) error) bool {
	add, ok := ctx.Value(afterCommitContextKey{}).(func(func(context.Context) error))
	if !ok {
		return false
	}
	add(fn)
	return true
}
