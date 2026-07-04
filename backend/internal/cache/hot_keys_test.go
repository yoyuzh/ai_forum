package cache

import "testing"

func TestHotScoreKeys(t *testing.T) {
	if got := HotPostCounterKey(42, HotCounterLike); got != "post:42:like_count" {
		t.Fatalf("like key = %q", got)
	}
	if got := HotPostScoreKey(42); got != "post:42:hot_score" {
		t.Fatalf("score key = %q", got)
	}
	if HotPostsZSetKey != "hot_posts:zset" || DirtyHotPostsSetKey != "dirty_hot_posts:set" {
		t.Fatalf("board keys = %q/%q", HotPostsZSetKey, DirtyHotPostsSetKey)
	}
}
