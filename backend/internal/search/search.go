// Package search provides the Elasticsearch client used by the search
// read-model. NewES constructs a client, pings the cluster, and verifies the
// IK Chinese analyzer is installed — IK absence fails readiness rather than
// warning, because ES without IK cannot do Chinese search and IK is an
// install-time requirement that cannot be rebuilt from MySQL.
package search

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/event"
)

// ikProbeText is the sample Chinese text sent to the _analyze endpoint to
// confirm the ik_smart analyzer is installed.
const ikProbeText = "中文测试"

const indexName = "forum_contents"
const syncConsumerName = "asynq.sync_search_index"

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

type SyncPayload struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	PostID    int64  `json:"post_id"`
	CommentID int64  `json:"comment_id,omitempty"`
	Title     string `json:"title,omitempty"`
	Status    string `json:"status,omitempty"`
}

type Document struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	PostID   int64  `json:"post_id"`
	AuthorID int64  `json:"author_id,omitempty"`
	Title    string `json:"title,omitempty"`
	Body     string `json:"body"`
	Status   string `json:"status,omitempty"`
}

type IndexStore interface {
	EnsureIndex(context.Context) error
	Upsert(context.Context, Document) error
	Delete(context.Context, string) error
}

type SyncHandler struct {
	db    database.DBTX
	store IndexStore
}

type indexedPost struct {
	ID       int64  `db:"id"`
	AuthorID int64  `db:"author_id"`
	Title    string `db:"title"`
	Content  string `db:"content"`
	Status   string `db:"status"`
}

type indexedComment struct {
	ID          int64          `db:"id"`
	PostID      int64          `db:"post_id"`
	UserID      sql.NullInt64  `db:"user_id"`
	CommentType string         `db:"comment_type"`
	AIAgentID   sql.NullInt64  `db:"ai_agent_id"`
	Content     string         `db:"content"`
	AgentName   sql.NullString `db:"agent_name"`
}

func NewSyncHandler(db database.DBTX, store IndexStore) *SyncHandler {
	return &SyncHandler{db: db, store: store}
}

func (h *SyncHandler) HandleSyncSearchIndex(ctx context.Context, body []byte) error {
	var payload SyncPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decode sync_search_index payload: %w", err)
	}
	if payload.EventID != "" {
		done, err := event.IsProcessed(ctx, h.db, payload.EventID, syncConsumerName)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
	if err := h.sync(ctx, payload); err != nil {
		return err
	}
	if payload.EventID != "" {
		return event.MarkProcessed(ctx, h.db, payload.EventID, syncConsumerName)
	}
	return nil
}

func (h *SyncHandler) sync(ctx context.Context, payload SyncPayload) error {
	switch payload.EventType {
	case "post.created", "post.updated":
		doc, err := h.BuildPostDocument(ctx, payload.PostID)
		if err != nil {
			return err
		}
		return h.store.Upsert(ctx, doc)
	case "comment.created", "ai.reply.completed":
		doc, err := h.BuildCommentDocument(ctx, payload.CommentID)
		if err != nil {
			return err
		}
		return h.store.Upsert(ctx, doc)
	case "post.deleted":
		return h.store.Delete(ctx, postDocID(payload.PostID))
	case "comment.deleted":
		return h.store.Delete(ctx, commentDocID(payload.CommentID))
	case "post.moderated":
		if payload.Status != "" && payload.Status != "NORMAL" {
			return h.store.Delete(ctx, postDocID(payload.PostID))
		}
		doc, err := h.BuildPostDocument(ctx, payload.PostID)
		if err != nil {
			return err
		}
		return h.store.Upsert(ctx, doc)
	case "ai.reply.failed":
		return nil
	default:
		return nil
	}
}

func (h *SyncHandler) BuildPostDocument(ctx context.Context, postID int64) (Document, error) {
	var p indexedPost
	err := h.db.GetContext(ctx, &p, `
		SELECT id, author_id, title, content, status
		FROM posts
		WHERE id = ? AND deleted_at IS NULL`, postID)
	if err != nil {
		return Document{}, err
	}
	return postDocument(p), nil
}

func (h *SyncHandler) BuildCommentDocument(ctx context.Context, commentID int64) (Document, error) {
	var c indexedComment
	err := h.db.GetContext(ctx, &c, `
		SELECT c.id, c.post_id, c.user_id, c.comment_type, c.ai_agent_id, c.content, a.name AS agent_name
		FROM comments c
		LEFT JOIN ai_agents a ON a.id = c.ai_agent_id
		WHERE c.id = ? AND c.deleted_at IS NULL`, commentID)
	if err != nil {
		return Document{}, err
	}
	return commentDocument(c), nil
}

func (h *SyncHandler) BuildAllDocuments(ctx context.Context) ([]Document, error) {
	var posts []indexedPost
	if err := h.db.SelectContext(ctx, &posts, `
		SELECT id, author_id, title, content, status
		FROM posts
		WHERE deleted_at IS NULL
		ORDER BY id`); err != nil {
		return nil, err
	}
	docs := make([]Document, 0, len(posts))
	for _, p := range posts {
		docs = append(docs, postDocument(p))
	}
	var comments []indexedComment
	if err := h.db.SelectContext(ctx, &comments, `
		SELECT c.id, c.post_id, c.user_id, c.comment_type, c.ai_agent_id, c.content, a.name AS agent_name
		FROM comments c
		LEFT JOIN ai_agents a ON a.id = c.ai_agent_id
		WHERE c.deleted_at IS NULL
		ORDER BY c.id`); err != nil {
		return nil, err
	}
	for _, c := range comments {
		docs = append(docs, commentDocument(c))
	}
	return docs, nil
}

func (h *SyncHandler) RebuildAll(ctx context.Context) error {
	if err := h.store.EnsureIndex(ctx); err != nil {
		return err
	}
	docs, err := h.BuildAllDocuments(ctx)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		if err := h.store.Upsert(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

func postDocument(p indexedPost) Document {
	return Document{ID: postDocID(p.ID), Type: "post", PostID: p.ID, AuthorID: p.AuthorID, Title: p.Title, Body: p.Content, Status: p.Status}
}

func commentDocument(c indexedComment) Document {
	docType := "comment"
	title := ""
	if c.CommentType == "AI" {
		docType = "ai_comment"
		title = c.AgentName.String
	}
	return Document{ID: commentDocID(c.ID), Type: docType, PostID: c.PostID, Title: title, Body: c.Content}
}

func postDocID(postID int64) string {
	return fmt.Sprintf("post:%d", postID)
}

func commentDocID(commentID int64) string {
	return fmt.Sprintf("comment:%d", commentID)
}

type ESIndexStore struct {
	client *es.Client
}

func NewESIndexStore(client *es.Client) *ESIndexStore {
	return &ESIndexStore{client: client}
}

func (s *ESIndexStore) EnsureIndex(ctx context.Context) error {
	body := `{"mappings":{"properties":{"id":{"type":"keyword"},"type":{"type":"keyword"},"post_id":{"type":"long"},"author_id":{"type":"long"},"title":{"type":"text","analyzer":"ik_smart"},"body":{"type":"text","analyzer":"ik_smart"},"status":{"type":"keyword"}}}}`
	res, err := s.client.Indices.Create(indexName, s.client.Indices.Create.WithBody(strings.NewReader(body)), s.client.Indices.Create.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 400 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(res.Body)
		if strings.Contains(buf.String(), "resource_already_exists_exception") {
			return nil
		}
		return fmt.Errorf("create es index: status %d: %s", res.StatusCode, buf.String())
	}
	if res.IsError() {
		return fmt.Errorf("create es index: status %d", res.StatusCode)
	}
	return nil
}

func (s *ESIndexStore) Upsert(ctx context.Context, doc Document) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	res, err := s.client.Index(indexName, bytes.NewReader(body), s.client.Index.WithDocumentID(doc.ID), s.client.Index.WithContext(ctx), s.client.Index.WithRefresh("true"))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("upsert es document: status %d", res.StatusCode)
	}
	return nil
}

func (s *ESIndexStore) Delete(ctx context.Context, id string) error {
	res, err := s.client.Delete(indexName, id, s.client.Delete.WithContext(ctx), s.client.Delete.WithRefresh("true"))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		_, _ = io.Copy(io.Discard, res.Body)
		return nil
	}
	if res.IsError() {
		return fmt.Errorf("delete es document: status %d", res.StatusCode)
	}
	return nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
