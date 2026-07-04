package post

import "time"

type Post struct {
	ID           int64     `db:"id" json:"id"`
	AuthorID     int64     `db:"author_id" json:"author_id"`
	Title        string    `db:"title" json:"title"`
	Content      string    `db:"content" json:"content"`
	Status       string    `db:"status" json:"status"`
	Tags         []string  `db:"-" json:"tags"`
	ViewCount    int64     `db:"view_count" json:"view_count"`
	CommentCount int64     `db:"comment_count" json:"comment_count"`
	LikeCount    int64     `db:"like_count" json:"like_count"`
	AIReplyCount int64     `db:"ai_reply_count" json:"ai_reply_count"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
