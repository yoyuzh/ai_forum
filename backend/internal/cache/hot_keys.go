package cache

import "fmt"

type HotCounter string

const (
	HotCounterView    HotCounter = "view"
	HotCounterLike    HotCounter = "like"
	HotCounterComment HotCounter = "comment"
	HotCounterAIReply HotCounter = "ai_reply"

	HotPostsZSetKey     = "hot_posts:zset"
	DirtyHotPostsSetKey = "dirty_hot_posts:set"
)

func HotPostCounterKey(postID int64, counter HotCounter) string {
	return fmt.Sprintf("post:%d:%s_count", postID, counter)
}

func HotPostScoreKey(postID int64) string {
	return fmt.Sprintf("post:%d:hot_score", postID)
}
