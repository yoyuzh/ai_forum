// Unit tests for the search package. These do NOT require a live Elasticsearch
// and run under the default `go test ./...` (no build tag). The stubbed HTTP
// transport simulates an ES without the IK plugin returning a 400.
package search

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRoundTripper returns a canned response for every request, letting us
// drive ikPresent without a live ES.
type stubRoundTripper struct {
	statusCode int
	body       string
	requests   []*http.Request
	bodies     []string
}

func (s *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		var b strings.Builder
		_, _ = io.Copy(&b, req.Body)
		s.bodies = append(s.bodies, b.String())
	}
	s.requests = append(s.requests, req)
	h := make(http.Header)
	// The ES client validates the X-Elastic-Product header on 2xx responses
	// (genuineCheckHeader). Without it the client rejects the response as
	// "not Elasticsearch". Set it so the stub behaves like a real ES for the
	// success path.
	h.Set("X-Elastic-Product", "Elasticsearch")
	return &http.Response{
		StatusCode: s.statusCode,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     h,
	}, nil
}

// newStubbedClient builds an *es.Client backed by a stub transport returning
// the given canned response for every HTTP request.
func newStubbedClient(t *testing.T, statusCode int, body string) *es.Client {
	t.Helper()
	transport := &stubRoundTripper{statusCode: statusCode, body: body}
	client, err := es.NewClient(es.Config{
		Addresses: []string{"http://stub:9200"},
		Transport: transport,
	})
	require.NoError(t, err)
	return client
}

// TestIKAbsentFailsReadiness simulates an ES without the IK plugin: the
// _analyze request returns HTTP 400 with a body naming the missing analyzer.
// ikPresent MUST return an error whose message begins with "es ik analyzer
// missing" (spec: elasticsearch-client, "IK absent fails readiness").
func TestIKAbsentFailsReadiness(t *testing.T) {
	client := newStubbedClient(t, 400,
		`{"error":{"root_cause":[{"type":"illegal_argument_exception","reason":"failed to find analyzer [ik_smart]"}]},"status":400}`)

	err := ikPresent(context.Background(), client)
	require.Error(t, err, "ikPresent must fail when the analyzer is missing")
	assert.True(t, strings.HasPrefix(err.Error(), "es ik analyzer missing"),
		"error must signal IK absence, got %q", err.Error())
	assert.Contains(t, err.Error(), "ik_smart")
}

// TestIKAbsentNon400FailsReadiness treats any non-200 error (here a 500) as
// IK-missing so readiness fails loudly rather than silently passing.
func TestIKAbsentNon400FailsReadiness(t *testing.T) {
	client := newStubbedClient(t, 500, `{"error":"server error","status":500}`)

	err := ikPresent(context.Background(), client)
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "es ik analyzer missing"),
		"non-200 error must also signal IK absence, got %q", err.Error())
}

// TestIKPresentSucceeds simulates an ES with IK: the _analyze request returns
// 200 with a token stream. ikPresent MUST return nil.
func TestIKPresentSucceeds(t *testing.T) {
	client := newStubbedClient(t, 200,
		`{"tokens":[{"token":"中","start_offset":0,"end_offset":1,"type":"CN_WORD","position":0}]}`)

	err := ikPresent(context.Background(), client)
	require.NoError(t, err, "ikPresent must succeed when IK returns 200")
}

// TestIsIKMissingBody covers the sentinel matching for a 400 body.
func TestIsIKMissingBody(t *testing.T) {
	assert.True(t, isIKMissingBody(`{"reason":"failed to find analyzer [ik_smart]"}`, 400))
	assert.True(t, isIKMissingBody(`{"reason":"analyzer not found"}`, 400))
	// 400 without a sentinel is still treated as missing (ambiguous = fail).
	assert.True(t, isIKMissingBody(`{"unrelated":"error"}`, 400))
	// Non-400 is treated as missing.
	assert.True(t, isIKMissingBody(`{}`, 500))
}

func TestEnsureIndexUsesForumContentsIKMappingAndIsIdempotent(t *testing.T) {
	transport := &stubRoundTripper{statusCode: 400, body: `{"error":{"type":"resource_already_exists_exception"}}`}
	client, err := es.NewClient(es.Config{Addresses: []string{"http://stub:9200"}, Transport: transport})
	require.NoError(t, err)
	store := NewESIndexStore(client)

	err = store.EnsureIndex(context.Background())

	require.NoError(t, err)
	require.Len(t, transport.requests, 1)
	assert.Equal(t, http.MethodPut, transport.requests[0].Method)
	assert.Contains(t, transport.requests[0].URL.Path, "forum_contents")
	assert.Contains(t, transport.bodies[0], `"analyzer":"ik_smart"`)
	assert.Contains(t, transport.bodies[0], `"type":{"type":"keyword"}`)
}

func TestESIndexStoreUpsertAndDeleteUseDocumentID(t *testing.T) {
	transport := &stubRoundTripper{statusCode: 200, body: `{}`}
	client, err := es.NewClient(es.Config{Addresses: []string{"http://stub:9200"}, Transport: transport})
	require.NoError(t, err)
	store := NewESIndexStore(client)
	doc := Document{ID: "post:42", Type: "post", PostID: 42, Title: "t", Body: "b"}

	require.NoError(t, store.Upsert(context.Background(), doc))
	require.NoError(t, store.Delete(context.Background(), doc.ID))

	require.Len(t, transport.requests, 2)
	assert.Equal(t, http.MethodPut, transport.requests[0].Method)
	assert.Contains(t, transport.requests[0].URL.Path, "post:42")
	assert.Equal(t, http.MethodDelete, transport.requests[1].Method)
	assert.Contains(t, transport.requests[1].URL.Path, "post:42")
	var got Document
	require.NoError(t, json.Unmarshal([]byte(transport.bodies[0]), &got))
	assert.Equal(t, doc, got)
}
