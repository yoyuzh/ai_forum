CREATE TABLE decision_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    post_id BIGINT NOT NULL,
    comment_id BIGINT NULL,
    ai_agent_id BIGINT NOT NULL,
    trigger_type VARCHAR(30) NOT NULL,
    willingness_score DECIMAL(5,4),
    threshold_value DECIMAL(5,4),
    decision VARCHAR(30) NOT NULL,
    reason VARCHAR(255),
    hit_tags JSON,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_decision_logs_post_agent_trigger (post_id, ai_agent_id, trigger_type),
    INDEX idx_decision_logs_post_id (post_id),
    INDEX idx_decision_logs_agent_id (ai_agent_id),
    CONSTRAINT fk_decision_logs_post FOREIGN KEY (post_id) REFERENCES posts(id),
    CONSTRAINT fk_decision_logs_agent FOREIGN KEY (ai_agent_id) REFERENCES ai_agents(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
