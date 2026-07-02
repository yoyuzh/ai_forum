package tagging

import (
	"context"
	"testing"
)

func TestRuleTaggerReturnsFiveTypes(t *testing.T) {
	tagger := RuleTagger{}

	tags := tagger.Tag(Post{ID: 42, Title: "Should we debate AI risk?", Content: "I am worried about safety"})

	seen := map[string]bool{}
	for _, tag := range tags {
		seen[tag.Type] = true
		if tag.Name == "" {
			t.Fatalf("empty tag name: %#v", tag)
		}
	}
	for _, want := range []string{"topic", "intent", "emotion", "debate", "risk"} {
		if !seen[want] {
			t.Fatalf("missing tag type %s in %#v", want, tags)
		}
	}
}

func TestHandlerWritesTagsAndAppendsPostTagged(t *testing.T) {
	posts := &recordingPostReader{post: Post{ID: 42, Title: "AI risk", Content: "Should we debate safety?"}}
	tags := &recordingTagWriter{}
	outbox := &recordingOutbox{}
	handler := NewHandler(posts, tags, outbox, RuleTagger{})

	if err := handler.HandleTagPost(context.Background(), 42); err != nil {
		t.Fatal(err)
	}

	if posts.postID != 42 {
		t.Fatalf("post read id = %d, want 42", posts.postID)
	}
	if len(tags.tags) != 5 {
		t.Fatalf("tags written = %d, want 5", len(tags.tags))
	}
	if outbox.eventType != "post.tagged" || outbox.aggregateID != 42 {
		t.Fatalf("outbox = %s/%d, want post.tagged/42", outbox.eventType, outbox.aggregateID)
	}
}

type recordingPostReader struct {
	postID int64
	post   Post
}

func (r *recordingPostReader) GetPost(_ context.Context, postID int64) (Post, error) {
	r.postID = postID
	return r.post, nil
}

type recordingTagWriter struct {
	postID int64
	tags   []Tag
}

func (w *recordingTagWriter) ReplaceTags(_ context.Context, postID int64, tags []Tag) error {
	w.postID = postID
	w.tags = append([]Tag(nil), tags...)
	return nil
}

type recordingOutbox struct {
	eventType   string
	aggregateID int64
}

func (o *recordingOutbox) AppendPostTagged(_ context.Context, postID int64, tags []Tag) error {
	o.eventType = "post.tagged"
	o.aggregateID = postID
	return nil
}
