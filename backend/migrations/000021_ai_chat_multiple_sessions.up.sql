ALTER TABLE ai_chat_sessions DROP INDEX uk_ai_chat_session_user_agent;
CREATE INDEX idx_ai_chat_sessions_user_agent_updated ON ai_chat_sessions (user_id, ai_agent_id, updated_at);
