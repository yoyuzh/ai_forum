// Package tagging generates post tags for AI decisions.
package tagging

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/event"
	forumpost "ai-forum/backend/internal/forum/post"
	forumtag "ai-forum/backend/internal/forum/tag"
	"ai-forum/backend/internal/outbox"
)

type Post struct {
	ID      int64
	Title   string
	Content string
}

type Tag struct {
	Type string
	Name string
}

type Tagger interface {
	Tag(Post) []Tag
}

type RuleTagger struct{}

type PostReader interface {
	GetPost(ctx context.Context, postID int64) (Post, error)
}

type TagWriter interface {
	ReplaceTags(ctx context.Context, postID int64, tags []Tag) error
}

type OutboxAppender interface {
	AppendPostTagged(ctx context.Context, postID int64, tags []Tag) error
}

type Handler struct {
	posts  PostReader
	tags   TagWriter
	outbox OutboxAppender
	tagger Tagger
}

type SQLHandler struct {
	db     *sqlx.DB
	tagger Tagger
}

func NewHandler(posts PostReader, tags TagWriter, outbox OutboxAppender, tagger Tagger) *Handler {
	return &Handler{posts: posts, tags: tags, outbox: outbox, tagger: tagger}
}

func NewSQLHandler(db *sqlx.DB, tagger Tagger) *SQLHandler {
	return &SQLHandler{db: db, tagger: tagger}
}

func (h *Handler) HandleTagPost(ctx context.Context, postID int64) error {
	post, err := h.posts.GetPost(ctx, postID)
	if err != nil {
		return err
	}
	tags := h.tagger.Tag(post)
	if err := h.tags.ReplaceTags(ctx, postID, tags); err != nil {
		return err
	}
	return h.outbox.AppendPostTagged(ctx, postID, tags)
}

func (h *SQLHandler) HandleTagPost(ctx context.Context, postID int64) error {
	postRepo := forumpost.NewSQLRepository()
	tagRepo := forumtag.NewSQLRepository()
	return database.RunInTx(ctx, h.db, func(tx *sqlx.Tx) error {
		p, err := postRepo.Get(ctx, tx, postID)
		if err != nil {
			return err
		}
		tags := h.tagger.Tag(Post{ID: p.ID, Title: p.Title, Content: p.Content})
		forumTags := make([]forumtag.Tag, 0, len(tags))
		for _, tag := range tags {
			forumTags = append(forumTags, forumtag.Tag{PostID: postID, Type: tag.Type, Name: tag.Name})
		}
		if err := tagRepo.Replace(ctx, tx, postID, forumTags); err != nil {
			return err
		}
		return outbox.Append(ctx, tx, outbox.Event{
			EventType:     event.PostTagged,
			AggregateType: "post",
			AggregateID:   postID,
			Payload:       map[string]any{"post_id": postID, "tags": tags},
		})
	})
}

func (RuleTagger) Tag(post Post) []Tag {
	text := strings.ToLower(post.Title + " " + post.Content)
	return []Tag{
		{Type: "topic", Name: topicTag(text)},
		{Type: "intent", Name: intentTag(text)},
		{Type: "emotion", Name: emotionTag(text)},
		{Type: "debate", Name: debateTag(text)},
		{Type: "risk", Name: riskTag(text)},
	}
}

func topicTag(text string) string {
	if strings.Contains(text, "ai") {
		return "ai"
	}
	if strings.Contains(text, "debate") {
		return "debate"
	}
	return "general"
}

func intentTag(text string) string {
	if strings.Contains(text, "?") || strings.Contains(text, "should") {
		return "question"
	}
	return "discussion"
}

func emotionTag(text string) string {
	if strings.Contains(text, "worried") || strings.Contains(text, "concern") {
		return "concerned"
	}
	return "neutral"
}

func debateTag(text string) string {
	if strings.Contains(text, "debate") || strings.Contains(text, "should") {
		return "high"
	}
	return "low"
}

func riskTag(text string) string {
	if strings.Contains(text, "risk") || strings.Contains(text, "safety") {
		return "sensitive"
	}
	return "normal"
}
