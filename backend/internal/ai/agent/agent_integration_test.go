//go:build integration

package agent

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

func TestRepositoryListsEnabledAgentsWithPreferences(t *testing.T) {
	db := newAgentIntegrationDB(t)
	repo := NewSQLRepository(db)

	agents, err := repo.ListEnabledWithPreferences(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(agents) < 3 {
		t.Fatalf("agents = %d, want at least 3", len(agents))
	}
	for _, agent := range agents {
		if agent.ID == 0 || agent.Name == "" || agent.ReplyThreshold == 0 {
			t.Fatalf("agent missing required fields: %#v", agent)
		}
		if !agent.AllowAutoReply {
			t.Fatalf("seeded agent should allow auto reply: %#v", agent)
		}
		if len(agent.Preferences) == 0 {
			t.Fatalf("agent %s has no preferences", agent.Name)
		}
	}
}

func newAgentIntegrationDB(t *testing.T) *sqlx.DB {
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
