package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DBTX is the common interface satisfied by both *sqlx.DB and *sqlx.Tx.
// Repositories depend on DBTX so the same code path runs against a plain
// connection or a caller-owned transaction — this is what makes inserting an
// outbox row on the same transaction as the business write natural (design D1).
// Verified at compile time by the assertions below.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Compile-time assertions that *sqlx.DB and *sqlx.Tx both satisfy DBTX.
var (
	_ DBTX = (*sqlx.DB)(nil)
	_ DBTX = (*sqlx.Tx)(nil)
)
