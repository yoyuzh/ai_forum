DROP INDEX idx_ai_chat_sessions_user_status_updated ON ai_chat_sessions;
DROP INDEX idx_ai_chat_sessions_status ON ai_chat_sessions;

ALTER TABLE ai_chat_messages
    DROP INDEX idx_ai_chat_messages_session_sequence,
    DROP INDEX uk_ai_chat_messages_request_id,
    DROP INDEX uk_ai_chat_messages_session_sequence,
    DROP COLUMN updated_at,
    DROP COLUMN model_name,
    DROP COLUMN request_id,
    DROP COLUMN sequence_no,
    DROP COLUMN status;

ALTER TABLE ai_chat_sessions
    DROP COLUMN message_count,
    DROP COLUMN last_message_preview,
    DROP COLUMN status;
