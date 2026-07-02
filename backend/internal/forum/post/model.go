package post

type Post struct {
	ID       int64  `db:"id" json:"id"`
	AuthorID int64  `db:"author_id" json:"author_id"`
	Title    string `db:"title" json:"title"`
	Content  string `db:"content" json:"content"`
	Status   string `db:"status" json:"status"`
}
