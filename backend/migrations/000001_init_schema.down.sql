-- 000001_init_schema.down.sql
-- Reverse of 000001. Only the baseline users table (owned by P1) is dropped.
-- Down migrations for outbox/processed_events live in their own files.
DROP TABLE IF EXISTS users;
