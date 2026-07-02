-- P4 extends the P1-owned users table; it must not recreate users.

ALTER TABLE users
    ADD COLUMN email VARCHAR(255) NULL AFTER username,
    ADD COLUMN display_name VARCHAR(80) NULL AFTER email,
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' AFTER role,
    ADD UNIQUE KEY uk_users_email (email),
    ADD INDEX idx_users_status (status);
