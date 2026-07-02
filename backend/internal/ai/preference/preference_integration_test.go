//go:build integration

package preference

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

func TestRepositoryListsByAgent(t *testing.T) {
	db := newPreferenceIntegrationDB(t)
	repo := NewSQLRepository(db)

	prefs, err := repo.ListByAgent(context.Background(), 1001)
	if err != nil {
		t.Fatal(err)
	}
	if len(prefs) == 0 {
		t.Fatal("expected seeded preferences")
	}
	for _, pref := range prefs {
		if pref.AgentID != 1001 || pref.TagType == "" || pref.TagName == "" || pref.Weight <= 0 {
			t.Fatalf("bad preference: %#v", pref)
		}
	}
}

func newPreferenceIntegrationDB(t *testing.T) *sqlx.DB {
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

	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, pass, host, port, name))
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
