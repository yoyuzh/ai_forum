-- 000001_init_schema.up.sql
-- P1 baseline schema. Establishes the utf8mb4/InnoDB `users` table that P1
-- seeds (000004) and that P4 extends via ALTER (000005_users no longer
-- CREATEs users — see openspec changes/p1 & p4). Outbox/processed_events live
-- in 000002/000003. Domain tables (posts/comments/...) are owned by P4.

CREATE TABLE users (
    id            BIGINT       NOT NULL AUTO_INCREMENT,
    username      VARCHAR(50)  NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    role          VARCHAR(20)  NOT NULL,
    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
