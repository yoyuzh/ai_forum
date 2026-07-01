package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RunInTx is the sole sanctioned primitive for multi-statement business+outbox
// writes. It begins a transaction, passes the *sqlx.Tx to fn, commits when fn
// returns nil, and rolls back when fn returns a non-nil error.
//
// If fn returns an error, the rollback is attempted and any rollback error is
// wrapped onto the original error so callers see both. RunInTx performs no MQ
// side effects — outbox rows written inside fn are published asynchronously by
// the outbox-publisher (database/AGENTS.md: "no MQ publish inside a tx").
func RunInTx(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// If commit succeeds we return nil. On any fn error we roll back.
	// committed is tracked so a panic-path rollback (via defer) does not mask a
	// successful commit.
	committed := false
	defer func() {
		if !committed {
			// Best-effort rollback; ignore ErrTxDone when the caller already
			// committed or rolled back inside fn (not expected, but safe).
			// On the panic path a deferred func cannot return a value to the
			// caller, so a rollback failure here is not surfaced — the panic
			// itself propagates. The fn-error path below wraps its rollback
			// error into the returned error; this defer only covers panics.
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				_ = rbErr
			}
		}
	}()

	if err := fn(tx); err != nil {
		// Roll back explicitly so the error is captured here, then wrap.
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			return fmt.Errorf("tx fn error (%w); rollback also failed: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true
	return nil
}
