package tag

type Tag struct {
	PostID int64
	Type   string
	Name   string
}

type HotTag struct {
	Name      string `db:"tag_name" json:"name"`
	PostCount int64  `db:"post_count" json:"post_count"`
}
