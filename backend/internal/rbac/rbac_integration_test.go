//go:build integration

package rbac

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func TestSQLXAdapterPersistsPolicies(t *testing.T) {
	db := newRBACIntegrationDB(t)
	authz, err := NewSQLXAuthorizer(DefaultModelPath(), db)
	if err != nil {
		t.Fatal(err)
	}
	if err := authz.AddPolicy("ADMIN", "post", "delete-any"); err != nil {
		t.Fatal(err)
	}

	reloaded, err := NewSQLXAuthorizer(DefaultModelPath(), db)
	if err != nil {
		t.Fatal(err)
	}
	allowed, err := reloaded.Enforce("ADMIN", "post", "delete-any")
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("persisted policy must allow after reload")
	}
}

func newRBACIntegrationDB(t *testing.T) *sqlx.DB {
	t.Helper()
	host := env("MYSQL_HOST", "127.0.0.1")
	port := env("MYSQL_PORT", "3306")
	user := env("MYSQL_USERNAME", "root")
	pass := env("MYSQL_PASSWORD", "ai_forum_root")
	name := env("MYSQL_DATABASE", "ai_forum")

	m, err := migrate.New("file://../../migrations", "mysql://"+user+":"+pass+"@tcp("+host+":"+port+")/"+name)
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
