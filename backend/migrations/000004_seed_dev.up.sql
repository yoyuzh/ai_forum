-- 000004_seed_dev.up.sql
-- DEV-ONLY seed. Inserts a single admin user with a known dev bcrypt hash.
-- This migration is NOT run in production (documented); production uses a
-- separate bootstrap path. The password is "admin123" (dev only) — bcrypt
-- cost 10. No API keys, JWT secrets, or tokens appear here (dev-seed-data
-- spec: "no real secrets in seed").
--
-- Fixed id=1 so the down migration can remove exactly this row without
-- affecting any other user. AI agent rows are NOT seeded here — P6 owns
-- ai_agents/ai_agent_tag_preferences and their seed (design D4).

INSERT INTO users (id, username, password_hash, role)
VALUES (
    1,
    'admin',
    '$2a$10$9LV7HawpWep1ITLI6JNmRuAnq4otsEhLpQ3LOCTiD6LU0pHy7ap9a',
    'ADMIN'
);
