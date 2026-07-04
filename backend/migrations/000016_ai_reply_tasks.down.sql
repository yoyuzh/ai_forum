DROP TABLE IF EXISTS ai_reply_tasks;

ALTER TABLE comments
    DROP COLUMN trigger_type;

ALTER TABLE posts
    DROP COLUMN ai_reply_count;
