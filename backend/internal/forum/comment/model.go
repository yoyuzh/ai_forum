package comment

type Comment struct {
	ID              int64
	PostID          int64
	UserID          int64
	ParentCommentID *int64
	CommentType     string
	Content         string
	Children        []*Comment
}
