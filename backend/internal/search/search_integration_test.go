//go:build integration

// Integration tests for the search package, run against a live Elasticsearch
// (with IK plugin) from docker-compose. Build tag `integration` keeps these
// out of the default `go test ./...` run.
//
// Run with:
//
//	ES_ADDRESSES=http://127.0.0.1:9200 \
//	go test -tags=integration ./internal/search/...
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/database"
)

// esCfgFromEnv builds the Elasticsearch config the same way the loader does,
// from the same env vars. Defaults match docker-compose so
// `docker compose up -d` + `go test -tags=integration` works out of the box.
func esCfgFromEnv() config.ElasticsearchConfig {
	if v := os.Getenv("ES_ADDRESSES"); v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return config.ElasticsearchConfig{Addresses: parts}
	}
	return config.ElasticsearchConfig{Addresses: []string{"http://127.0.0.1:9200"}}
}

// TestESPingAndIK verifies NewES succeeds against a docker-compose ES that
// has the IK plugin installed (spec: elasticsearch-client, "IK present").
func TestESPingAndIK(t *testing.T) {
	cfg := esCfgFromEnv()

	client, err := NewES(cfg)
	require.NoError(t, err, "NewES must connect and verify IK against live ES")
	assert.NotNil(t, client)
}

func TestSyncSearchIndexWritesAndDeletesPostInES(t *testing.T) {
	ctx := context.Background()
	db, m := databaseTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	client, err := NewES(esCfgFromEnv())
	require.NoError(t, err)
	store := NewESIndexStore(client)
	require.NoError(t, store.EnsureIndex(ctx))
	handler := NewSyncHandler(db, store)

	res, err := db.ExecContext(ctx, `INSERT INTO posts (author_id, title, content, status) VALUES (1, 'P9 search title', 'P9 search body', 'NORMAL')`)
	require.NoError(t, err)
	postID, err := res.LastInsertId()
	require.NoError(t, err)

	body, err := json.Marshal(SyncPayload{EventID: "evt-p9-search-live", EventType: "post.created", PostID: postID})
	require.NoError(t, err)
	require.NoError(t, handler.HandleSyncSearchIndex(ctx, body))

	waitForESHit(t, client, "P9 search title", true)

	body, err = json.Marshal(SyncPayload{EventID: "evt-p9-search-delete-live", EventType: "post.deleted", PostID: postID})
	require.NoError(t, err)
	require.NoError(t, handler.HandleSyncSearchIndex(ctx, body))
	waitForESHit(t, client, "P9 search title", false)
}

func waitForESHit(t *testing.T, client *es.Client, term string, want bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	query := `{"query":{"match":{"title":` + strconv.Quote(term) + `}}}`
	for time.Now().Before(deadline) {
		res, err := client.Search(client.Search.WithIndex(indexName), client.Search.WithBody(strings.NewReader(query)))
		if err == nil {
			var body struct {
				Hits struct {
					Total struct {
						Value int `json:"value"`
					} `json:"total"`
				} `json:"hits"`
			}
			if decodeErr := json.NewDecoder(res.Body).Decode(&body); decodeErr == nil {
				_ = res.Body.Close()
				if (body.Hits.Total.Value > 0) == want {
					return
				}
			} else {
				_ = res.Body.Close()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("ES hit for %q did not become %v", term, want)
}

func databaseTestDB(t *testing.T) (*sqlx.DB, *migrate.Migrate) {
	t.Helper()
	cfg := mysqlCfgFromEnv()
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

func mysqlCfgFromEnv() config.MySQLConfig {
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
