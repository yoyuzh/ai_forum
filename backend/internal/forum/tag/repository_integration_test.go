//go:build integration

package tag

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func TestSQLRepositoryStoresAndListsTags(t *testing.T) {
	db := newTagIntegrationDB(t)
	postID := seedPost(t, db)
	repo := NewSQLRepository()
	ctx := context.Background()

	if err := repo.Replace(ctx, db, postID, []Tag{{Type: "topic", Name: "go"}, {Type: "risk", Name: "spam"}}); err != nil {
		t.Fatal(err)
	}
	tags, err := repo.List(ctx, db, postID)
	if err != nil {
		t.Fatal(err)
	}
	grouped := GroupByType(tags)
	if grouped["topic"][0] != "go" || grouped["risk"][0] != "spam" {
		t.Fatalf("grouped = %#v", grouped)
	}
}

func seedPost(t *testing.T, db *sqlx.DB) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO posts (author_id, title, content, status) VALUES (1, 'p', 'body', 'NORMAL')`)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func newTagIntegrationDB(t *testing.T) *sqlx.DB {
	t.Helper()
	host := env("MYSQL_HOST", "127.0.0.1")
	port := env("MYSQL_PORT", "3306")
	user := env("MYSQL_USERNAME", "root")
	pass := env("MYSQL_PASSWORD", "ai_forum_root")
	name := env("MYSQL_DATABASE", "ai_forum")
	m, err := migrate.New("file://../../../migrations", "mysql://"+user+":"+pass+"@tcp("+host+":"+port+")/"+name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = m.Close() })
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
