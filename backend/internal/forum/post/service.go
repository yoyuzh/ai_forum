package post

import (
	"context"
	"fmt"
	"strings"

	"ai-forum/backend/internal/outbox"
)

type CreateInput struct {
	AuthorID int64
	Title    string
	Content  string
}

type UpdateInput struct {
	PostID   int64
	AuthorID int64
	Title    string
	Content  string
}

type AppendFunc func(context.Context, DBTX, outbox.Event) error

type Service struct {
	repo   Repository
	append AppendFunc
}

func NewService(repo Repository, append AppendFunc) *Service {
	return &Service{repo: repo, append: append}
}

func (s *Service) CreatePost(ctx context.Context, tx DBTX, in CreateInput) (Post, error) {
	title := strings.TrimSpace(in.Title)
	content := strings.TrimSpace(in.Content)
	if in.AuthorID <= 0 || title == "" || content == "" {
		return Post{}, fmt.Errorf("invalid post")
	}
	p, err := s.repo.Create(ctx, tx, Post{
		AuthorID: in.AuthorID,
		Title:    title,
		Content:  content,
		Status:   "NORMAL",
	})
	if err != nil {
		return Post{}, err
	}
	if err := s.append(ctx, tx, outbox.Event{
		EventType:     "post.created",
		AggregateType: "post",
		AggregateID:   p.ID,
		Payload:       map[string]any{"post_id": p.ID, "author_id": p.AuthorID},
	}); err != nil {
		return Post{}, err
	}
	return p, nil
}

func (s *Service) List(ctx context.Context, tx DBTX) ([]Post, error) {
	return s.repo.List(ctx, tx)
}

func (s *Service) Get(ctx context.Context, tx DBTX, postID int64) (Post, error) {
	if postID <= 0 {
		return Post{}, fmt.Errorf("invalid post")
	}
	return s.repo.Get(ctx, tx, postID)
}

func (s *Service) UpdateOwn(ctx context.Context, tx DBTX, in UpdateInput) (Post, error) {
	title := strings.TrimSpace(in.Title)
	content := strings.TrimSpace(in.Content)
	if in.PostID <= 0 || in.AuthorID <= 0 || title == "" || content == "" {
		return Post{}, fmt.Errorf("invalid post update")
	}
	p, err := s.repo.Update(ctx, tx, Post{ID: in.PostID, AuthorID: in.AuthorID, Title: title, Content: content})
	if err != nil {
		return Post{}, err
	}
	if err := s.append(ctx, tx, outbox.Event{
		EventType:     "post.updated",
		AggregateType: "post",
		AggregateID:   p.ID,
		Payload:       map[string]any{"post_id": p.ID, "author_id": p.AuthorID},
	}); err != nil {
		return Post{}, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, tx DBTX, postID int64) error {
	if postID <= 0 {
		return fmt.Errorf("invalid post delete")
	}
	if err := s.repo.SoftDelete(ctx, tx, postID); err != nil {
		return err
	}
	return s.append(ctx, tx, outbox.Event{
		EventType:     "post.deleted",
		AggregateType: "post",
		AggregateID:   postID,
		Payload:       map[string]any{"post_id": postID},
	})
}

func (s *Service) UpdateStatus(ctx context.Context, tx DBTX, postID int64, status string) error {
	status = strings.TrimSpace(status)
	if postID <= 0 || status == "" {
		return fmt.Errorf("invalid post status update")
	}
	if err := s.repo.UpdateStatus(ctx, tx, postID, status); err != nil {
		return err
	}
	return s.append(ctx, tx, outbox.Event{
		EventType:     "post.moderated",
		AggregateType: "post",
		AggregateID:   postID,
		Payload:       map[string]any{"post_id": postID, "status": status},
	})
}
