CREATE TABLE post_tags (
    id BIGINT NOT NULL AUTO_INCREMENT,
    post_id BIGINT NOT NULL,
    tag_type VARCHAR(32) NOT NULL,
    tag_name VARCHAR(80) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_post_tags_post_type_name (post_id, tag_type, tag_name),
    INDEX idx_post_tags_type_name (tag_type, tag_name),
    CONSTRAINT fk_post_tags_post FOREIGN KEY (post_id) REFERENCES posts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
