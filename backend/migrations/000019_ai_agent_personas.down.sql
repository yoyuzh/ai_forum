DELETE FROM ai_agent_tag_preferences WHERE ai_agent_id BETWEEN 1001 AND 1012;
DELETE FROM ai_agents WHERE id BETWEEN 1004 AND 1012;

UPDATE ai_agents SET
    name = 'cohere_observer',
    enabled = TRUE,
    reply_threshold = 0.6000,
    activity_level = 0.5000,
    allow_auto_reply = TRUE,
    allow_mention = TRUE,
    allow_followup = TRUE,
    is_fallback = TRUE
WHERE id = 1001;

UPDATE ai_agents SET
    name = 'debate_synthesizer',
    enabled = TRUE,
    reply_threshold = 0.6200,
    activity_level = 0.6500,
    allow_auto_reply = TRUE,
    allow_mention = TRUE,
    allow_followup = TRUE,
    is_fallback = FALSE
WHERE id = 1002;

UPDATE ai_agents SET
    name = 'risk_moderator',
    enabled = TRUE,
    reply_threshold = 0.5800,
    activity_level = 0.5500,
    allow_auto_reply = TRUE,
    allow_mention = TRUE,
    allow_followup = TRUE,
    is_fallback = FALSE
WHERE id = 1003;

INSERT INTO ai_agent_tag_preferences
    (ai_agent_id, tag_type, tag_name, weight)
VALUES
    (1001, 'topic', 'general', 0.5000),
    (1001, 'intent', 'discussion', 0.6000),
    (1002, 'topic', 'debate', 0.9000),
    (1002, 'debate', 'high', 0.9000),
    (1003, 'risk', 'sensitive', 0.9000),
    (1003, 'emotion', 'concerned', 0.7000);

ALTER TABLE ai_agents DROP COLUMN system_prompt;
