DROP INDEX idx_ai_chat_sessions_user_agent_updated ON ai_chat_sessions;
ALTER TABLE ai_chat_sessions ADD UNIQUE KEY uk_ai_chat_session_user_agent (user_id, ai_agent_id);
