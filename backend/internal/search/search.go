// Package search provides the Elasticsearch client used by the search
// read-model. NewES constructs a client, pings the cluster, and verifies the
// IK Chinese analyzer is installed — IK absence fails readiness rather than
// warning, because ES without IK cannot do Chinese search and IK is an
// install-time requirement that cannot be rebuilt from MySQL.
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"ai-forum/backend/internal/config"
)

// ikProbeText is the sample Chinese text sent to the _analyze endpoint to
// confirm the ik_smart analyzer is installed.
const ikProbeText = "中文测试"

// ikAnalyzerMissingSentinel is a substring of the ES error body that indicates
// the analyzer could not be resolved. ES returns HTTP 400 with a body like:
//
//	{"error":{"root_cause":[{"type":"illegal_argument_exception","reason":"failed to find analyzer [ik_smart]"}]}}
//
// Matching on this substring keeps the probe independent of the exact JSON
// shape across ES versions.
var ikAnalyzerMissingSentinels = []string{
	"failed to find analyzer",
	"analyzer [ik_smart]",
	"analyzer not found",
}

// NewES constructs an Elasticsearch client from cfg, pings the cluster, and
// probes the IK analyzer. It returns an error if the ping fails or if the IK
// analyzer is missing — both are readiness failures, not warnings.
func NewES(cfg config.ElasticsearchConfig) (*es.Client, error) {
	client, err := es.NewClient(es.Config{Addresses: cfg.Addresses})
	if err != nil {
		return nil, fmt.Errorf("es new client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("es ping: %w", err)
	}
	if res.IsError() {
		_ = res.Body.Close()
		return nil, fmt.Errorf("es ping: status %d", res.StatusCode)
	}
	_ = res.Body.Close()

	if err := ikPresent(ctx, client); err != nil {
		return nil, err
	}

	return client, nil
}

// ikPresent issues an _analyze request with the ik_smart analyzer on a sample
// Chinese text. It returns an error wrapping the underlying failure when the
// request fails, the response is an error, or the response body indicates the
// analyzer is missing. The error message starts with "es ik analyzer missing"
// so callers can distinguish the readiness-gate failure from other errors.
func ikPresent(ctx context.Context, client *es.Client) error {
	body := map[string]string{
		"analyzer": "ik_smart",
		"text":     ikProbeText,
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("es ik analyzer missing: marshal body: %w", err)
	}

	req := esapi.IndicesAnalyzeRequest{
		Body: bytes.NewReader(buf),
	}
	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("es ik analyzer missing: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// Read the body to inspect for the analyzer-missing signal. On a
		// healthy ES with IK, this branch is never hit. On ES without IK,
		// ES returns 400 with a body naming the missing analyzer.
		var buf bytes.Buffer
		if _, readErr := buf.ReadFrom(res.Body); readErr != nil {
			return fmt.Errorf("es ik analyzer missing: status %d (body read: %v)", res.StatusCode, readErr)
		}
		respBody := buf.String()
		if isIKMissingBody(respBody, res.StatusCode) {
			return fmt.Errorf("es ik analyzer missing: status %d: %s", res.StatusCode, respBody)
		}
		return fmt.Errorf("es ik analyzer missing: status %d: %s", res.StatusCode, respBody)
	}

	return nil
}

// isIKMissingBody reports whether the response body and status code indicate
// the ik_smart analyzer is absent. ES returns HTTP 400 with a body containing
// "failed to find analyzer" (or a similar phrase) when the analyzer is not
// installed. A non-400 error body is treated as IK-missing to be safe — the
// probe text and analyzer are the only variables.
func isIKMissingBody(body string, statusCode int) bool {
	if statusCode != 400 {
		return true
	}
	lower := strings.ToLower(body)
	for _, sentinel := range ikAnalyzerMissingSentinels {
		if strings.Contains(lower, strings.ToLower(sentinel)) {
			return true
		}
	}
	// 400 without a recognized sentinel is ambiguous; treat as missing so the
	// operator sees a readiness failure rather than a silent pass.
	return true
}
