//go:build integration

package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/auth"
	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/database"
)

func TestPostCreateSucceedsWithoutElasticsearchClient(t *testing.T) {
	db, m := p9BootstrapTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 18080},
		JWT:    config.JWTConfig{Secret: "p9-chaos-secret", ExpireHours: 1},
	}
	app := &App{Cfg: cfg, DB: db}
	process := app.NewAPIServer()
	httpProcess, ok := process.(*HTTPProcess)
	require.True(t, ok)
	token, err := auth.NewTokenManager(cfg.JWT.Secret, time.Hour).Issue(auth.Subject{UserID: 1, Username: "admin", Role: "ADMIN"})
	require.NoError(t, err)
	title := fmt.Sprintf("es down %d", time.Now().UnixNano())

	req := httptest.NewRequest(http.MethodPost, "/api/posts", strings.NewReader(fmt.Sprintf(`{"title":%q,"content":"mysql still writes"}`, title)))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	httpProcess.srv.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var rows int
	require.NoError(t, db.GetContext(context.Background(), &rows, `SELECT COUNT(*) FROM posts WHERE title = ?`, title))
	require.Equal(t, 1, rows)
}

func p9BootstrapTestDB(t *testing.T) (*sqlx.DB, *migrate.Migrate) {
	t.Helper()
	cfg := p9MySQLCfgFromEnv()
	db, err := database.NewMySQL(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	wd, err := os.Getwd()
	require.NoError(t, err)
	src := "file://" + filepath.Join(wd, "..", "..", "migrations")
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	m, err := migrate.New(src, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })
	return db, m
}

func p9MySQLCfgFromEnv() config.MySQLConfig {
	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}
	port := 3306
	if v := os.Getenv("MYSQL_PORT"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &port); err != nil {
			port = 3306
		}
	}
	return config.MySQLConfig{
		Host:     get("MYSQL_HOST", "127.0.0.1"),
		Port:     port,
		Username: get("MYSQL_USERNAME", "root"),
		Password: get("MYSQL_PASSWORD", "ai_forum_root"),
		Database: get("MYSQL_DATABASE", "ai_forum"),
	}
}
