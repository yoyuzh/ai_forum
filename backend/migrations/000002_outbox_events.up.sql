-- 000002_outbox_events.up.sql
-- Verbatim architecture §8.4. This table is owned by P1; no later migration
-- SHALL CREATE or DROP it. Later phases that need columns SHALL ALTER under a
-- new migration number and document the dependency (db-migrations spec,
-- single-owner requirement).

CREATE TABLE outbox_events (
    id           BIGINT       PRIMARY KEY AUTO_INCREMENT,
    event_id     VARCHAR(64)  NOT NULL UNIQUE,
    event_type   VARCHAR(100) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id BIGINT       NOT NULL,
    payload      JSON         NOT NULL,
    status       VARCHAR(20)  DEFAULT 'PENDING',
    retry_count  INT          DEFAULT 0,
    created_at   DATETIME,
    published_at DATETIME,
    INDEX idx_outbox_status_created_at (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
