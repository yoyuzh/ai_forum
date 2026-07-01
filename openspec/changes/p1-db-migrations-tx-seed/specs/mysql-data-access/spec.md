## ADDED Requirements

### Requirement: MySQL connection with correct charset
`NewMySQL(cfg config.MySQL) (*sqlx.DB, error)` SHALL return a sqlx DB using a DSN with `parseTime=true`, `loc=Local`, `charset=utf8mb4`, `collation=utf8mb4_unicode_ci`, and a connection timeout. It SHALL set sane `MaxOpenConns`, `MaxIdleConns`, and `ConnMaxLifetime` defaults.

#### Scenario: Connection pings successfully
- **WHEN** `NewMySQL` is called against a running MySQL 8.4 container with valid credentials
- **THEN** the returned `*sqlx.DB` passes `PingContext`

### Requirement: RunInTx transaction primitive
`RunInTx(ctx, db, fn func(tx *sqlx.Tx) error) error` SHALL begin a transaction, execute `fn` with the `*sqlx.Tx`, commit when `fn` returns nil, and roll back when `fn` returns a non-nil error. Any rollback error SHALL be wrapped and returned alongside the original error.

#### Scenario: Commit on success
- **WHEN** `fn` inserts a row and returns nil
- **THEN** the row is visible after `RunInTx` returns

#### Scenario: Rollback on error
- **WHEN** `fn` inserts a row then returns a non-nil error
- **THEN** the inserted row is NOT present after `RunInTx` returns

### Requirement: Repository DBTX interface
The database package SHALL define a `DBTX` interface satisfied by both `*sqlx.DB` and `*sqlx.Tx` so that repositories execute against either without code changes, enabling outbox rows to be inserted on the same transaction as business writes.

#### Scenario: Repository works under tx
- **WHEN** a repository method is called with a `*sqlx.Tx` (via `DBTX`)
- **THEN** its writes participate in the caller's transaction and roll back if the caller errors
