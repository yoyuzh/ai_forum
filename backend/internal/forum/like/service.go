package like

import (
	"context"
	"fmt"

	"ai-forum/backend/internal/outbox"
)

type Repository interface {
	Like(context.Context, DBTX, int64, int64) (bool, error)
	Unlike(context.Context, DBTX, int64, int64) (bool, error)
	Count(context.Context, DBTX, int64) (int, error)
}

type AppendFunc func(context.Context, DBTX, outbox.Event) error

type Service struct {
	repo   Repository
	append AppendFunc
}

func NewService(repo Repository, append AppendFunc) *Service {
	return &Service{repo: repo, append: append}
}

func (s *Service) Like(ctx context.Context, tx DBTX, userID, postID int64) error {
	if userID <= 0 || postID <= 0 {
		return fmt.Errorf("invalid like")
	}
	changed, err := s.repo.Like(ctx, tx, userID, postID)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return s.append(ctx, tx, outbox.Event{
		EventType:     "post.liked",
		AggregateType: "post",
		AggregateID:   postID,
		Payload:       map[string]any{"post_id": postID, "user_id": userID},
	})
}

func (s *Service) Unlike(ctx context.Context, tx DBTX, userID, postID int64) error {
	if userID <= 0 || postID <= 0 {
		return fmt.Errorf("invalid unlike")
	}
	changed, err := s.repo.Unlike(ctx, tx, userID, postID)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return s.append(ctx, tx, outbox.Event{
		EventType:     "post.unliked",
		AggregateType: "post",
		AggregateID:   postID,
		Payload:       map[string]any{"post_id": postID, "user_id": userID},
	})
}

func (s *Service) Count(ctx context.Context, tx DBTX, postID int64) (int, error) {
	if postID <= 0 {
		return 0, fmt.Errorf("invalid post id")
	}
	return s.repo.Count(ctx, tx, postID)
}
