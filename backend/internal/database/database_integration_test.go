//go:build integration

// Integration tests for the database package, run against a live MySQL 8.4
// container (docker-compose up -d mysql). Build tag `integration` keeps these
// out of the default `go test ./...` run.
//
// Run with:
//
//	MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 MYSQL_USERNAME=root \
//	MYSQL_PASSWORD=ai_forum_root MYSQL_DATABASE=ai_forum \
//	go test -tags=integration ./internal/database/...
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/config"
)

// mysqlCfgFromEnv builds the MySQL config the same way the loader does, from
// the same env vars. Defaults match docker-compose so `docker compose up -d`
// + `go test -tags=integration` works out of the box.
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

// migrationsPath resolves backend/migrations relative to the test file.
func migrationsPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// internal/database -> backend/migrations
	abs, err := filepath.Abs(filepath.Join(wd, "..", "..", "migrations"))
	if err != nil {
		return "", err
	}
	return "file://" + abs, nil
}

// newTestDB returns a *sqlx.DB plus a migrate instance wired to the
// migrations directory. The caller is responsible for running migrations up
// and down.
func newTestDB(t *testing.T) (*sqlx.DB, *migrate.Migrate) {
	t.Helper()
	cfg := mysqlCfgFromEnv()

	db, err := NewMySQL(cfg)
	require.NoError(t, err, "NewMySQL must connect to the live MySQL container")
	t.Cleanup(func() { _ = db.Close() })

	src, err := migrationsPath()
	require.NoError(t, err)
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	m, err := migrate.New(src, dsn)
	require.NoError(t, err, "migrate.New must parse migrations dir + DSN")
	t.Cleanup(func() { _, _ = m.Close() })

	return db, m
}

// TestMySQLConnection verifies NewMySQL produces a pingable connection
// (spec: mysql-data-access, "Connection pings successfully").
func TestMySQLConnection(t *testing.T) {
	db, _ := newTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, db.PingContext(ctx), "PingContext must succeed against live MySQL")
}

// TestMigrationsApplyAndReverse verifies `migrate-up` applies all migrations
// cleanly on a fresh DB and `migrate-down` reverses them (spec: db-migrations,
// "Fresh apply and reverse").
func TestMigrationsApplyAndReverse(t *testing.T) {
	_, m := newTestDB(t)

	// Start from a clean slate: force down to 0, then up.
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Down baseline failed: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}

	// Down must reverse cleanly.
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Down failed: %v", err)
	}
}

// TestRunInTxCommit verifies a row written inside RunInTx with a nil return
// is persisted (spec: mysql-data-access, "Commit on success").
func TestRunInTxCommit(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()

	err := RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO outbox_events (event_id, event_type, aggregate_type, aggregate_id, payload, status, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, NOW())`,
			"evt-commit-1", "post.created", "post", 1, `{"v":1}`, "PENDING")
		return err
	})
	require.NoError(t, err)

	var count int
	require.NoError(t, db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM outbox_events WHERE event_id = ?`, "evt-commit-1"))
	assert.Equal(t, 1, count, "committed row must be visible after RunInTx")
}

// TestRunInTxRollback verifies a row written inside RunInTx before a non-nil
// return is NOT persisted (spec: mysql-data-access, "Rollback on error").
func TestRunInTxRollback(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()

	err := RunInTx(ctx, db, func(tx *sqlx.Tx) error {
		if _, e := tx.ExecContext(ctx,
			`INSERT INTO outbox_events (event_id, event_type, aggregate_type, aggregate_id, payload, status, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, NOW())`,
			"evt-rollback-1", "post.created", "post", 2, `{"v":1}`, "PENDING"); e != nil {
			return e
		}
		// Return a non-nil error to trigger rollback.
		return fmt.Errorf("simulated business error")
	})
	require.Error(t, err)

	var count int
	require.NoError(t, db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM outbox_events WHERE event_id = ?`, "evt-rollback-1"))
	assert.Equal(t, 0, count, "rolled-back row must not be visible after RunInTx")
}

// TestOutboxSchemaMatchesSpec asserts outbox_events columns/index match §8.4
// via information_schema (spec: db-migrations, "Schema introspection matches").
func TestOutboxSchemaMatchesSpec(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()

	type col struct {
		ColumnName string `db:"COLUMN_NAME"`
		DataType   string `db:"DATA_TYPE"`
		IsNullable string `db:"IS_NULLABLE"`
	}
	var cols []col
	require.NoError(t, db.SelectContext(ctx, &cols, `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE
		FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'outbox_events'
		ORDER BY ordinal_position`))

	colByName := map[string]col{}
	for _, c := range cols {
		colByName[c.ColumnName] = c
	}
	for _, name := range []string{
		"id", "event_id", "event_type", "aggregate_type", "aggregate_id",
		"payload", "status", "retry_count", "created_at", "published_at",
	} {
		_, ok := colByName[name]
		assert.True(t, ok, "outbox_events must have column %s", name)
	}
	assert.Equal(t, "bigint", colByName["id"].DataType)
	assert.Equal(t, "varchar", colByName["event_id"].DataType)
	assert.Equal(t, "json", colByName["payload"].DataType)
	assert.Equal(t, "NO", colByName["event_id"].IsNullable, "event_id is NOT NULL")

	// idx_outbox_status_created_at must exist (composite index → count by name).
	var idxCount int
	require.NoError(t, db.GetContext(ctx, &idxCount, `
		SELECT COUNT(DISTINCT index_name) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = 'outbox_events'
		  AND index_name = 'idx_outbox_status_created_at'`))
	assert.Equal(t, 1, idxCount, "idx_outbox_status_created_at must exist")
}

// TestProcessedEventsSchemaMatchesSpec asserts processed_events matches §9.2
// and the unique key rejects duplicates (spec: db-migrations).
func TestProcessedEventsSchemaMatchesSpec(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()

	var idxCount int
	require.NoError(t, db.GetContext(ctx, &idxCount, `
		SELECT COUNT(DISTINCT index_name) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = 'processed_events'
		  AND index_name = 'uk_processed_event_consumer'`))
	assert.Equal(t, 1, idxCount, "uk_processed_event_consumer must exist")

	// First insert succeeds.
	_, err := db.ExecContext(ctx,
		`INSERT INTO processed_events (event_id, consumer_name, processed_at) VALUES (?, ?, NOW())`,
		"evt-dup-1", "worker-tag")
	require.NoError(t, err)

	// Duplicate (event_id, consumer_name) must fail.
	_, err = db.ExecContext(ctx,
		`INSERT INTO processed_events (event_id, consumer_name, processed_at) VALUES (?, ?, NOW())`,
		"evt-dup-1", "worker-tag")
	require.Error(t, err, "duplicate (event_id, consumer_name) must violate the unique key")
}

func TestAIDomainSchemaAndSeedMatchesP6(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()

	for table, columns := range map[string][]string{
		"ai_agents": {
			"id", "name", "enabled", "reply_threshold", "activity_level",
			"allow_auto_reply", "allow_mention", "allow_followup", "is_fallback",
		},
		"ai_agent_tag_preferences": {
			"id", "ai_agent_id", "tag_type", "tag_name", "weight",
		},
		"decision_logs": {
			"id", "post_id", "comment_id", "ai_agent_id", "trigger_type",
			"willingness_score", "threshold_value", "decision", "reason", "hit_tags", "created_at",
		},
	} {
		for _, column := range columns {
			var count int
			require.NoError(t, db.GetContext(ctx, &count, `
				SELECT COUNT(*) FROM information_schema.columns
				WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`, table, column))
			assert.Equal(t, 1, count, "%s.%s must exist", table, column)
		}
	}

	var enabledAgents int
	require.NoError(t, db.GetContext(ctx, &enabledAgents, `SELECT COUNT(*) FROM ai_agents WHERE enabled = TRUE`))
	assert.GreaterOrEqual(t, enabledAgents, 3, "P6 dev seed must include at least 3 enabled agents")

	var fallbackAgents int
	require.NoError(t, db.GetContext(ctx, &fallbackAgents, `SELECT COUNT(*) FROM ai_agents WHERE is_fallback = TRUE`))
	assert.GreaterOrEqual(t, fallbackAgents, 1, "P6 dev seed must include a fallback observer")

	var preferenceCount int
	require.NoError(t, db.GetContext(ctx, &preferenceCount, `SELECT COUNT(*) FROM ai_agent_tag_preferences`))
	assert.GreaterOrEqual(t, preferenceCount, 5, "P6 dev seed must include tag preferences")
}

func TestAIReplyTasksSchemaMatchesP7(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()

	for _, column := range []string{
		"id", "post_id", "parent_comment_id", "parent_comment_id_norm", "ai_agent_id",
		"trigger_type", "status", "attempt_count", "last_error", "comment_id",
		"created_at", "updated_at",
	} {
		var count int
		require.NoError(t, db.GetContext(ctx, &count, `
			SELECT COUNT(*) FROM information_schema.columns
			WHERE table_schema = DATABASE() AND table_name = 'ai_reply_tasks' AND column_name = ?`, column))
		assert.Equal(t, 1, count, "ai_reply_tasks.%s must exist", column)
	}

	var generation string
	require.NoError(t, db.GetContext(ctx, &generation, `
		SELECT generation_expression
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		  AND table_name = 'ai_reply_tasks'
		  AND column_name = 'parent_comment_id_norm'`))
	assert.Contains(t, generation, "coalesce(`parent_comment_id`,0)")

	var uniqueColumns []string
	require.NoError(t, db.SelectContext(ctx, &uniqueColumns, `
		SELECT column_name
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		  AND table_name = 'ai_reply_tasks'
		  AND index_name = 'uk_ai_reply_task'
		ORDER BY seq_in_index`))
	assert.Equal(t, []string{"post_id", "parent_comment_id_norm", "ai_agent_id", "trigger_type"}, uniqueColumns)

	var aiReplyCountColumns int
	require.NoError(t, db.GetContext(ctx, &aiReplyCountColumns, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'posts' AND column_name = 'ai_reply_count'`))
	assert.Equal(t, 1, aiReplyCountColumns, "posts.ai_reply_count must exist for P7 counters")

	var viewCountColumns int
	require.NoError(t, db.GetContext(ctx, &viewCountColumns, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'posts' AND column_name = 'view_count'`))
	assert.Equal(t, 1, viewCountColumns, "posts.view_count must exist for P10 hot score snapshots")

	var triggerTypeColumns int
	require.NoError(t, db.GetContext(ctx, &triggerTypeColumns, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'comments' AND column_name = 'trigger_type'`))
	assert.Equal(t, 1, triggerTypeColumns, "comments.trigger_type must exist for P7 AI replies")
}

func TestNotificationsSchemaMatchesP9(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()

	for _, column := range []string{"id", "recipient_id", "type", "payload", "read_at", "created_at"} {
		var count int
		require.NoError(t, db.GetContext(ctx, &count, `
			SELECT COUNT(*) FROM information_schema.columns
			WHERE table_schema = DATABASE() AND table_name = 'notifications' AND column_name = ?`, column))
		assert.Equal(t, 1, count, "notifications.%s must exist", column)
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO notifications (recipient_id, type, payload)
		VALUES (?, ?, JSON_OBJECT('post_id', ?))`, int64(1), "ai.reply.completed", int64(42))
	require.NoError(t, err)
}

func TestAIReplyTasksConcurrentInsertTreatsConflictAsIdempotentSuccess(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up failed: %v", err)
	}
	ctx := context.Background()

	postID := seedP7Post(t, ctx, db)

	insert := func() error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO ai_reply_tasks (post_id, parent_comment_id, ai_agent_id, trigger_type, status)
			VALUES (?, NULL, ?, ?, ?)`,
			postID, int64(1001), "AUTO", "PENDING")
		if isDuplicateKey(err) {
			return nil
		}
		return err
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- insert()
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	var rows int
	require.NoError(t, db.GetContext(ctx, &rows, `
		SELECT COUNT(*) FROM ai_reply_tasks
		WHERE post_id = ? AND parent_comment_id_norm = 0 AND ai_agent_id = ? AND trigger_type = ?`,
		postID, int64(1001), "AUTO"))
	assert.Equal(t, 1, rows)

	var failed int
	require.NoError(t, db.GetContext(ctx, &failed, `
		SELECT COUNT(*) FROM ai_reply_tasks
		WHERE post_id = ? AND status = 'FAILED'`, postID))
	assert.Equal(t, 0, failed)
}

func seedP7Post(t *testing.T, ctx context.Context, db *sqlx.DB) int64 {
	t.Helper()
	res, err := db.ExecContext(ctx, `
		INSERT INTO posts (author_id, title, content, status)
		VALUES (?, ?, ?, ?)`, 1, "p7 post", "content", "NORMAL")
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return id
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 ||
		err != nil && strings.Contains(err.Error(), "Duplicate entry")
}

// TestSeedDevAdmin verifies 000004 leaves exactly one admin user and no AI
// rows (AI tables do not exist yet — P6 owns them), and that down reverses
// cleanly (spec: dev-seed-data).
func TestSeedDevAdmin(t *testing.T) {
	db, m := newTestDB(t)
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Down baseline failed: %v", err)
	}
	if err := m.Migrate(4); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Migrate(4) failed: %v", err)
	}
	ctx := context.Background()

	var adminCount int
	require.NoError(t, db.GetContext(ctx, &adminCount,
		`SELECT COUNT(*) FROM users WHERE role = 'ADMIN'`))
	assert.Equal(t, 1, adminCount, "exactly one dev admin user must be seeded")

	var adminID int
	require.NoError(t, db.GetContext(ctx, &adminID,
		`SELECT id FROM users WHERE username = 'admin'`))
	assert.Equal(t, 1, adminID, "seeded admin must have fixed id=1")

	// Password is a bcrypt hash ($2a$ / $2b$), not plaintext.
	var hash string
	require.NoError(t, db.GetContext(ctx, &hash,
		`SELECT password_hash FROM users WHERE id = 1`))
	assert.True(t, len(hash) > 4 && (hash[:4] == "$2a$" || hash[:4] == "$2b$"),
		"admin password must be a bcrypt hash, got %q", hash)

	// AI tables must not exist (owned by P6).
	var tableCount int
	require.NoError(t, db.GetContext(ctx, &tableCount, `
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name IN ('ai_agents', 'ai_agent_tag_preferences')`))
	assert.Equal(t, 0, tableCount, "no AI seed tables may exist in P1")

	// Down reverses the seed cleanly (only the admin row is removed).
	prevVersion, dirty, err := m.Version()
	require.NoError(t, err, "m.Version before down")
	require.False(t, dirty)

	// Roll back only the seed migration (step down by 1).
	require.NoError(t, m.Steps(-1))

	var remainingAdmins int
	require.NoError(t, db.GetContext(ctx, &remainingAdmins,
		`SELECT COUNT(*) FROM users WHERE username = 'admin'`))
	assert.Equal(t, 0, remainingAdmins, "seed admin must be removed by 000004 down")

	// Restore state so other subtests/Cleanup find a consistent DB.
	_ = prevVersion
	require.NoError(t, m.Steps(1))
}

// TestNoLiteralSecretsInMigrations is the task-6.4 grep test: no migration
// file may contain literal JWT/INTERNAL_API/AI keys, and the admin password
// must be a bcrypt hash, not plaintext (spec: dev-seed-data).
func TestNoLiteralSecretsInMigrations(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := filepath.Join(wd, "..", "..", "migrations")
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		require.NoError(t, err)
		body := string(data)
		for _, bad := range []string{"JWT_SECRET", "INTERNAL_API_TOKEN", "AI_API_KEY", "change-me"} {
			assert.NotContains(t, body, bad, "%s must not contain %s", e.Name(), bad)
		}
	}
}

// Compile-time: keep the sql import used if expanded later.
var _ = sql.ErrTxDone
