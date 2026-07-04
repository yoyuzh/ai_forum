package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"

	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/event"
)

type EventEnvelope struct {
	EventID   string         `json:"EventID"`
	EventType string         `json:"EventType"`
	Payload   map[string]any `json:"Payload"`
}

type Enqueuer interface {
	Enqueue(ctx context.Context, taskType string, payload any) error
}

type optionEnqueuer interface {
	EnqueueWithOptions(ctx context.Context, taskType string, payload any, opts ...asynq.Option) error
}

type AsynqEnqueuer struct {
	client *asynq.Client
}

func NewAsynqEnqueuer(client *asynq.Client) *AsynqEnqueuer {
	return &AsynqEnqueuer{client: client}
}

func (e *AsynqEnqueuer) Enqueue(ctx context.Context, taskType string, payload any) error {
	return e.EnqueueWithOptions(ctx, taskType, payload)
}

func (e *AsynqEnqueuer) EnqueueWithOptions(ctx context.Context, taskType string, payload any, opts ...asynq.Option) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal task payload: %w", err)
	}
	_, err = e.client.EnqueueContext(ctx, asynq.NewTask(taskType, body), opts...)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue %s: %w", taskType, err)
	}
	return nil
}

type PostCreatedConsumer struct {
	enqueuer     Enqueuer
	processed    ProcessedStore
	consumerName string
	taskType     string
}

type ProcessedStore interface {
	IsProcessed(context.Context, string, string) (bool, error)
	MarkProcessed(context.Context, string, string) error
}

type SQLProcessedStore struct {
	db database.DBTX
}

func NewSQLProcessedStore(db database.DBTX) *SQLProcessedStore {
	return &SQLProcessedStore{db: db}
}

func (s *SQLProcessedStore) IsProcessed(ctx context.Context, eventID, consumerName string) (bool, error) {
	return event.IsProcessed(ctx, s.db, eventID, consumerName)
}

func (s *SQLProcessedStore) MarkProcessed(ctx context.Context, eventID, consumerName string) error {
	return event.MarkProcessed(ctx, s.db, eventID, consumerName)
}

type ConsumerOption func(*PostCreatedConsumer)

func WithProcessedStore(store ProcessedStore, consumerName string) ConsumerOption {
	return func(c *PostCreatedConsumer) {
		c.processed = store
		c.consumerName = consumerName
	}
}

func NewPostCreatedConsumer(enqueuer Enqueuer, opts ...ConsumerOption) *PostCreatedConsumer {
	c := &PostCreatedConsumer{enqueuer: enqueuer, taskType: TagPost}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func NewPostTaggedConsumer(enqueuer Enqueuer, opts ...ConsumerOption) *PostCreatedConsumer {
	c := &PostCreatedConsumer{enqueuer: enqueuer, taskType: DecideAIReply}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func NewSearchIndexConsumer(enqueuer Enqueuer, opts ...ConsumerOption) *PostCreatedConsumer {
	c := &PostCreatedConsumer{enqueuer: enqueuer, taskType: SyncSearchIndex}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func NewNotificationConsumer(enqueuer Enqueuer, opts ...ConsumerOption) *PostCreatedConsumer {
	c := &PostCreatedConsumer{enqueuer: enqueuer, taskType: SendNotification}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *PostCreatedConsumer) Handle(ctx context.Context, body []byte) error {
	var env EventEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("decode post.created event: %w", err)
	}
	shouldMark := false
	if c.processed != nil && env.EventID != "" {
		done, err := c.processed.IsProcessed(ctx, env.EventID, c.consumerName)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		shouldMark = true
	}
	postID, ok := numberAsInt64(env.Payload["post_id"])
	if !ok {
		return fmt.Errorf("%s missing post_id", env.EventType)
	}
	payload := any(TagPostPayload{PostID: postID})
	switch c.taskType {
	case DecideAIReply:
		payload = DecideAIReplyPayload{PostID: postID}
	case SyncSearchIndex:
		payload = SyncSearchIndexPayload{
			EventID:         env.EventID,
			EventType:       env.EventType,
			PostID:          postID,
			CommentID:       payloadInt64(env.Payload, "comment_id"),
			MentionedUserID: payloadInt64(env.Payload, "mentioned_user_id"),
			Title:           payloadString(env.Payload, "title"),
			Status:          payloadString(env.Payload, "status"),
		}
	case SendNotification:
		payload = SendNotificationPayload{
			EventID:         env.EventID,
			EventType:       env.EventType,
			PostID:          postID,
			CommentID:       payloadInt64(env.Payload, "comment_id"),
			MentionedUserID: payloadInt64(env.Payload, "mentioned_user_id"),
		}
	}
	if err := c.enqueuer.Enqueue(ctx, c.taskType, payload); err != nil {
		return err
	}
	if shouldMark {
		return c.processed.MarkProcessed(ctx, env.EventID, c.consumerName)
	}
	return nil
}

func payloadInt64(payload map[string]any, key string) int64 {
	n, _ := numberAsInt64(payload[key])
	return n
}

func payloadString(payload map[string]any, key string) string {
	s, _ := payload[key].(string)
	return s
}

func numberAsInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), n > 0
	case int64:
		return n, n > 0
	case int:
		return int64(n), n > 0
	default:
		return 0, false
	}
}
