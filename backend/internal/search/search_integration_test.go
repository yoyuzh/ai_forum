//go:build integration

// Integration tests for the search package, run against a live Elasticsearch
// (with IK plugin) from docker-compose. Build tag `integration` keeps these
// out of the default `go test ./...` run.
//
// Run with:
//
//	ES_ADDRESSES=http://127.0.0.1:9200 \
//	go test -tags=integration ./internal/search/...
package search

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/config"
)

// esCfgFromEnv builds the Elasticsearch config the same way the loader does,
// from the same env vars. Defaults match docker-compose so
// `docker compose up -d` + `go test -tags=integration` works out of the box.
func esCfgFromEnv() config.ElasticsearchConfig {
	if v := os.Getenv("ES_ADDRESSES"); v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return config.ElasticsearchConfig{Addresses: parts}
	}
	return config.ElasticsearchConfig{Addresses: []string{"http://127.0.0.1:9200"}}
}

// TestESPingAndIK verifies NewES succeeds against a docker-compose ES that
// has the IK plugin installed (spec: elasticsearch-client, "IK present").
func TestESPingAndIK(t *testing.T) {
	cfg := esCfgFromEnv()

	client, err := NewES(cfg)
	require.NoError(t, err, "NewES must connect and verify IK against live ES")
	assert.NotNil(t, client)
}
