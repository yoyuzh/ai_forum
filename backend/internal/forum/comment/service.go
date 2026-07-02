package comment

import (
	"context"
	"fmt"
	"strings"

	"ai-forum/backend/internal/outbox"
)

type CreateInput struct {
	PostID          int64
	UserID          int64
	ParentCommentID *int64
	Content         string
}

type AppendFunc func(context.Context, DBTX, outbox.Event) error

type Service struct {
	repo   Repository
	append AppendFunc
}

func NewService(repo Repository, append AppendFunc) *Service {
	return &Service{repo: repo, append: append}
}

func (s *Service) Create(ctx context.Context, tx DBTX, in CreateInput) (Comment, error) {
	content := strings.TrimSpace(in.Content)
	if in.PostID <= 0 || in.UserID <= 0 || content == "" {
		return Comment{}, fmt.Errorf("invalid comment")
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
	if err := s.repo.IncrementCommentCount(ctx, tx, in.PostID); err != nil {
		return Comment{}, err
	}
	if err := s.append(ctx, tx, outbox.Event{
		EventType:     "comment.created",
		AggregateType: "comment",
		AggregateID:   c.ID,
		Payload:       map[string]any{"comment_id": c.ID, "post_id": c.PostID},
	}); err != nil {
		return Comment{}, err
	}
	return c, nil
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
