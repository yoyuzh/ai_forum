-- 000003_processed_events.up.sql
-- Verbatim architecture §9.2. Owned by P1; no later migration SHALL CREATE or
-- DROP it (db-migrations spec, single-owner requirement).

CREATE TABLE processed_events (
    id            BIGINT      PRIMARY KEY AUTO_INCREMENT,
    event_id      VARCHAR(64)  NOT NULL,
    consumer_name VARCHAR(100) NOT NULL,
    processed_at  DATETIME,
    UNIQUE KEY uk_processed_event_consumer (event_id, consumer_name),
    INDEX idx_processed_events_processed_at (processed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
