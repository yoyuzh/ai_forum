package comment

type Comment struct {
	ID              int64   `db:"id" json:"id"`
	PostID          int64   `db:"post_id" json:"post_id"`
	UserID          int64   `db:"user_id" json:"user_id"`
	ParentCommentID *int64  `db:"parent_comment_id" json:"parent_comment_id"`
	CommentType     string  `db:"comment_type" json:"comment_type"`
	AIAgentID       *int64  `db:"ai_agent_id" json:"ai_agent_id"`
	TriggerType     string  `db:"trigger_type" json:"trigger_type,omitempty"`
	Content         string  `db:"content" json:"content"`
	Author          *Author `db:"-" json:"author,omitempty"`
	Children        []*Comment
}

type Author struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar,omitempty"`
	IsAI     bool   `json:"isAi"`
	Role     string `json:"role,omitempty"`
}

type MentionAgent struct {
	ID           int64  `db:"id"`
	Name         string `db:"name"`
	Enabled      bool   `db:"enabled"`
	AllowMention bool   `db:"allow_mention"`
}

type CommentMention struct {
	CommentID int64
	AIAgentID int64
}
