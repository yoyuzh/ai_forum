package favorite

import (
	"context"
	"fmt"

	"ai-forum/backend/internal/outbox"
)

type Repository interface {
	Favorite(context.Context, DBTX, int64, int64) (bool, error)
	Unfavorite(context.Context, DBTX, int64, int64) (bool, error)
	List(context.Context, DBTX, int64) ([]int64, error)
}

type AppendFunc func(context.Context, DBTX, outbox.Event) error

type Service struct {
	repo   Repository
	append AppendFunc
}

func NewService(repo Repository, append AppendFunc) *Service {
	return &Service{repo: repo, append: append}
}

func (s *Service) Favorite(ctx context.Context, tx DBTX, userID, postID int64) error {
	if userID <= 0 || postID <= 0 {
		return fmt.Errorf("invalid favorite")
	}
	changed, err := s.repo.Favorite(ctx, tx, userID, postID)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return s.append(ctx, tx, outbox.Event{
		EventType:     "post.favorited",
		AggregateType: "post",
		AggregateID:   postID,
		Payload:       map[string]any{"post_id": postID, "user_id": userID},
	})
}

func (s *Service) Unfavorite(ctx context.Context, tx DBTX, userID, postID int64) error {
	if userID <= 0 || postID <= 0 {
		return fmt.Errorf("invalid unfavorite")
	}
	changed, err := s.repo.Unfavorite(ctx, tx, userID, postID)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return s.append(ctx, tx, outbox.Event{
		EventType:     "post.unfavorited",
		AggregateType: "post",
		AggregateID:   postID,
		Payload:       map[string]any{"post_id": postID, "user_id": userID},
	})
}

func (s *Service) List(ctx context.Context, tx DBTX, userID int64) ([]int64, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	return s.repo.List(ctx, tx, userID)
}
