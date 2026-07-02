CREATE TABLE ai_agent_tag_preferences (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    ai_agent_id BIGINT NOT NULL,
    tag_type VARCHAR(30) NOT NULL,
    tag_name VARCHAR(50) NOT NULL,
    weight DECIMAL(5,4) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_agent_tag (ai_agent_id, tag_type, tag_name),
    INDEX idx_ai_agent_tag_preferences_agent (ai_agent_id),
    CONSTRAINT fk_ai_agent_tag_preferences_agent FOREIGN KEY (ai_agent_id) REFERENCES ai_agents(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO ai_agent_tag_preferences
    (ai_agent_id, tag_type, tag_name, weight)
VALUES
    (1001, 'topic', 'general', 0.5000),
    (1001, 'intent', 'discussion', 0.6000),
    (1002, 'topic', 'debate', 0.9000),
    (1002, 'debate', 'high', 0.9000),
    (1003, 'risk', 'sensitive', 0.9000),
    (1003, 'emotion', 'concerned', 0.7000);
