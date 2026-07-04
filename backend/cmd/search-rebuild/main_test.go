package main

import (
	"testing"

	"ai-forum/backend/internal/config"
)

func TestValidateSearchRebuildConfigRequiresOnlyMySQLAndES(t *testing.T) {
	cfg := &config.Config{
		MySQL: config.MySQLConfig{
			Host:     "127.0.0.1",
			Port:     3306,
			Username: "root",
			Password: "pw",
			Database: "ai_forum",
		},
		Elasticsearch: config.ElasticsearchConfig{Addresses: []string{"http://127.0.0.1:9200"}},
	}

	if err := validateSearchRebuildConfig(cfg); err != nil {
		t.Fatalf("validateSearchRebuildConfig() error = %v, want nil", err)
	}
}
