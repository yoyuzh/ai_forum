// Package database owns the MySQL connection, the transaction primitive, and
// the DBTX interface that lets repositories execute against either a
// *sqlx.DB or a *sqlx.Tx. It performs no MQ side effects (database/AGENTS.md).
package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql" // registered for side effects

	"ai-forum/backend/internal/config"
)

// Pool defaults — chosen to fit a single api-server instance behind a MySQL
// 8.4 container; revisited if throughput demands it.
const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 5 * time.Minute
	defaultConnTimeout     = 10 * time.Second
)

// NewMySQL returns a *sqlx.DB configured for utf8mb4 with sane pool defaults.
// The DSN forces parseTime (so DATETIME maps to time.Time), loc=Local,
// utf8mb4 charset/collation, and a 10s connect timeout (spec: mysql-data-access).
func NewMySQL(cfg config.MySQLConfig) (*sqlx.DB, error) {
	dsn := buildDSN(cfg)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql %s:%d/%s: %w",
			cfg.Host, cfg.Port, cfg.Database, err)
	}
	db.SetMaxOpenConns(defaultMaxOpenConns)
	db.SetMaxIdleConns(defaultMaxIdleConns)
	db.SetConnMaxLifetime(defaultConnMaxLifetime)
	return db, nil
}

// buildDSN constructs the MySQL DSN in go-sql-driver format:
// user:pass@tcp(host:port)/db?params
func buildDSN(cfg config.MySQLConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=%s",
		cfg.Username, cfg.Password,
		cfg.Host, cfg.Port,
		cfg.Database,
		defaultConnTimeout,
	)
}
