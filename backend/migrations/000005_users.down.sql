ALTER TABLE users
    DROP INDEX idx_users_status,
    DROP INDEX uk_users_email,
    DROP COLUMN status,
    DROP COLUMN display_name,
    DROP COLUMN email;
