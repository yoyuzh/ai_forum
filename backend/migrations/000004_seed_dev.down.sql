-- 000004_seed_dev.down.sql
-- Remove ONLY the seeded dev admin row (fixed id=1, username='admin') so the
-- migration is safe to re-run and never deletes real users.
DELETE FROM users WHERE id = 1 AND username = 'admin';
