CREATE TABLE ai_agents (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    reply_threshold DECIMAL(5,4) NOT NULL DEFAULT 0.6000,
    activity_level DECIMAL(5,4) NOT NULL DEFAULT 0.5000,
    allow_auto_reply BOOLEAN NOT NULL DEFAULT TRUE,
    allow_mention BOOLEAN NOT NULL DEFAULT TRUE,
    allow_followup BOOLEAN NOT NULL DEFAULT TRUE,
    is_fallback BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ai_agents_enabled (enabled),
    INDEX idx_ai_agents_fallback (is_fallback)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO ai_agents
    (id, name, enabled, reply_threshold, activity_level, allow_auto_reply, allow_mention, allow_followup, is_fallback)
VALUES
    (1001, 'cohere_observer', TRUE, 0.6000, 0.5000, TRUE, TRUE, TRUE, TRUE),
    (1002, 'debate_synthesizer', TRUE, 0.6200, 0.6500, TRUE, TRUE, TRUE, FALSE),
    (1003, 'risk_moderator', TRUE, 0.5800, 0.5500, TRUE, TRUE, TRUE, FALSE);
