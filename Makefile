# AI Forum — top-level Makefile.
# Migration targets read MySQL connection details from the same env vars the
# backend config loader uses (architecture §14.1), so CI and local share one
# path. MYSQL_DSN may be set directly to override; otherwise it is built from
# MYSQL_HOST/PORT/USERNAME/PASSWORD/DATABASE (db-migrations spec).
#
# golang-migrate is invoked via `go run` against the pinned CLI source so no
# global install is required. The `mysql://` DSN scheme is what the migrate
# CLI expects.

# golang-migrate is invoked via a local wrapper (`backend/cmd/migrate`) that
# blank-imports the mysql + file drivers. Running the upstream CLI via
# `go run github.com/.../cmd/migrate@v4.18.1` does not compile the driver
# subpackages, so `mysql://` is "unknown driver" — the wrapper fixes that.
MIGRATE_CMD := go -C backend run ./cmd/migrate
MIGRATIONS_DIR := $(abspath backend/migrations)

# Build MYSQL_DSN (mysql://user:pass@tcp(host:port)/db) from components unless
# the caller supplied MYSQL_DSN directly.
MYSQL_DSN ?= mysql://$(MYSQL_USERNAME):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

.PHONY: help migrate-up migrate-down migrate-create migrate-force

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-18s %s\n", $$1, $$2}'

migrate-up: ## Apply all pending migrations
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(MYSQL_DSN)" up

migrate-down: ## Roll back the most recent migration
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(MYSQL_DSN)" down

migrate-create: ## Create a new migration pair: make migrate-create NAME=add_column
	@test -n "$(NAME)" || { echo "usage: make migrate-create NAME=add_column"; exit 1; }
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) create $(NAME)

migrate-force: ## Force a migration version: make migrate-force V=3
	@test -n "$(V)" || { echo "usage: make migrate-force V=3"; exit 1; }
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(MYSQL_DSN)" force $(V)
