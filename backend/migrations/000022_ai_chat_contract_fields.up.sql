ALTER TABLE ai_chat_sessions
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    ADD COLUMN last_message_preview VARCHAR(255) NULL,
    ADD COLUMN message_count INT NOT NULL DEFAULT 0;

ALTER TABLE ai_chat_messages
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'DONE',
    ADD COLUMN sequence_no INT NULL,
    ADD COLUMN request_id VARCHAR(64) NULL,
    ADD COLUMN model_name VARCHAR(100) NULL,
    ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

UPDATE ai_chat_messages m
JOIN (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY id) AS rn
    FROM ai_chat_messages
) x ON x.id = m.id
SET m.sequence_no = x.rn;

UPDATE ai_chat_sessions s
LEFT JOIN (
    SELECT session_id, COUNT(*) AS message_count
    FROM ai_chat_messages
    GROUP BY session_id
) c ON c.session_id = s.id
LEFT JOIN (
    SELECT session_id, content
    FROM (
        SELECT session_id, content, ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY sequence_no DESC) AS rn
        FROM ai_chat_messages
    ) ranked
    WHERE rn = 1
) last_msg ON last_msg.session_id = s.id
SET
    s.message_count = COALESCE(c.message_count, 0),
    s.last_message_preview = LEFT(COALESCE(last_msg.content, ''), 255);

ALTER TABLE ai_chat_messages
    MODIFY COLUMN sequence_no INT NOT NULL,
    ADD UNIQUE KEY uk_ai_chat_messages_session_sequence (session_id, sequence_no),
    ADD UNIQUE KEY uk_ai_chat_messages_request_id (request_id),
    ADD INDEX idx_ai_chat_messages_session_sequence (session_id, sequence_no);

CREATE INDEX idx_ai_chat_sessions_status ON ai_chat_sessions (status);
CREATE INDEX idx_ai_chat_sessions_user_status_updated ON ai_chat_sessions (user_id, status, updated_at);
