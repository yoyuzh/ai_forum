ALTER TABLE posts
    ADD COLUMN ai_reply_count INT NOT NULL DEFAULT 0 AFTER comment_count;

ALTER TABLE comments
    ADD COLUMN trigger_type VARCHAR(30) NULL AFTER ai_agent_id;

CREATE TABLE ai_reply_tasks (
    id BIGINT NOT NULL AUTO_INCREMENT,
    post_id BIGINT NOT NULL,
    parent_comment_id BIGINT NULL,
    parent_comment_id_norm BIGINT GENERATED ALWAYS AS (COALESCE(parent_comment_id,0)) STORED,
    ai_agent_id BIGINT NOT NULL,
    trigger_type VARCHAR(30) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    attempt_count INT NOT NULL DEFAULT 0,
    last_error VARCHAR(255) NULL,
    comment_id BIGINT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_ai_reply_task (post_id, parent_comment_id_norm, ai_agent_id, trigger_type),
    INDEX idx_ai_reply_tasks_post_status (post_id, status),
    INDEX idx_ai_reply_tasks_agent_id (ai_agent_id),
    INDEX idx_ai_reply_tasks_comment_id (comment_id),
    CONSTRAINT fk_ai_reply_tasks_post FOREIGN KEY (post_id) REFERENCES posts(id),
    CONSTRAINT fk_ai_reply_tasks_parent_comment FOREIGN KEY (parent_comment_id) REFERENCES comments(id),
    CONSTRAINT fk_ai_reply_tasks_agent FOREIGN KEY (ai_agent_id) REFERENCES ai_agents(id),
    CONSTRAINT fk_ai_reply_tasks_comment FOREIGN KEY (comment_id) REFERENCES comments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
