CREATE TABLE casbin_rule (
    id BIGINT NOT NULL AUTO_INCREMENT,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100) NULL,
    v1 VARCHAR(100) NULL,
    v2 VARCHAR(100) NULL,
    v3 VARCHAR(100) NULL,
    v4 VARCHAR(100) NULL,
    v5 VARCHAR(100) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
