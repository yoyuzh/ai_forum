ALTER TABLE ai_chat_messages DROP CHECK chk_ai_chat_messages_role;

UPDATE ai_chat_messages
SET role = CASE role
    WHEN 'user' THEN 'USER'
    WHEN 'assistant' THEN 'AI'
    ELSE role
END;

ALTER TABLE ai_chat_messages
    ADD CONSTRAINT chk_ai_chat_messages_role CHECK (role IN ('USER', 'AI'));
