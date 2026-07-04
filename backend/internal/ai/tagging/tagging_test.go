package tagging

import (
	"context"
	"testing"
)

func TestRuleTaggerReturnsFiveTypes(t *testing.T) {
	tagger := RuleTagger{}

	tags := tagger.Tag(context.Background(), Post{ID: 42, Title: "Should we debate AI risk?", Content: "I am worried about safety"})

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

func TestParseTagsFiltersInvalidModelOutput(t *testing.T) {
	got := ParseTags("```json\n{\"topic\":[\"学习规划\",\"求助\"],\"intent\":[\"求建议\",\"求助\"],\"emotion\":[\"焦虑\"],\"debate\":[\"争议性低\"],\"risk\":[\"正常\",\"未知\"]}\n```")

	if len(got.Topic) != 1 || got.Topic[0] != "学习规划" {
		t.Fatalf("topic = %#v", got.Topic)
	}
	if len(got.Intent) != 1 || got.Intent[0] != "求建议" {
		t.Fatalf("intent = %#v", got.Intent)
	}
	if len(got.Risk) != 1 || got.Risk[0] != "正常" {
		t.Fatalf("risk = %#v", got.Risk)
	}
}

func TestParseTagsFallsBackOnInvalidJSON(t *testing.T) {
	got := ParseTags("not json")

	if len(got.Intent) != 1 || got.Intent[0] != "求建议" || len(got.Debate) != 1 || got.Debate[0] != "争议性低" || len(got.Risk) != 1 || got.Risk[0] != "正常" {
		t.Fatalf("fallback tags = %#v", got)
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
