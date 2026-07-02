// Package event defines RabbitMQ domain event contracts and consumer idempotency.
package event

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"ai-forum/backend/internal/database"
)

const (
	PostCreated      = "post.created"
	PostUpdated      = "post.updated"
	PostDeleted      = "post.deleted"
	CommentCreated   = "comment.created"
	CommentDeleted   = "comment.deleted"
	PostTagged       = "post.tagged"
	AIReplyCompleted = "ai.reply.completed"
	AIReplyFailed    = "ai.reply.failed"
	PostModerated    = "post.moderated"
	UserMentioned    = "user.mentioned"
)

// Envelope is the architecture §7.5 event payload shape.
type Envelope struct {
	EventID       string    `json:"eventId"`
	EventType     string    `json:"eventType"`
	AggregateType string    `json:"aggregateType"`
	AggregateID   int64     `json:"aggregateId"`
	OccurredAt    time.Time `json:"occurredAt"`
	Payload       any       `json:"payload"`
}

// Contract records phase ownership for the P5 contract-ownership gate.
type ContractInfo struct {
	PublisherOwner string
	ConsumerOwner  string
}

var contracts = map[string]ContractInfo{
	PostCreated:      {PublisherOwner: "P4 forum/post", ConsumerOwner: "P6 tag_post, P9 search/notification"},
	PostUpdated:      {PublisherOwner: "P4 forum/post", ConsumerOwner: "P9 search/notification"},
	PostDeleted:      {PublisherOwner: "P4 forum/post", ConsumerOwner: "P9 search/notification"},
	CommentCreated:   {PublisherOwner: "P4 forum/comment", ConsumerOwner: "P7/P8/P9 workers"},
	CommentDeleted:   {PublisherOwner: "P4 forum/comment", ConsumerOwner: "P9 search sync"},
	PostTagged:       {PublisherOwner: "P6 tag_post worker", ConsumerOwner: "P6 decide_ai_reply"},
	AIReplyCompleted: {PublisherOwner: "P7 generate_ai_reply", ConsumerOwner: "P9 notification/search"},
	AIReplyFailed:    {PublisherOwner: "P7 generate_ai_reply", ConsumerOwner: "P9 notification"},
	PostModerated:    {PublisherOwner: "P4 admin moderation", ConsumerOwner: "P9 search sync"},
	UserMentioned:    {PublisherOwner: "P8 mention parser", ConsumerOwner: "P8/P9 workers"},
}

// Types returns the documented §8.5 domain event types in stable order.
func Types() []string {
	return []string{
		PostCreated,
		PostUpdated,
		PostDeleted,
		CommentCreated,
		CommentDeleted,
		PostTagged,
		AIReplyCompleted,
		AIReplyFailed,
		PostModerated,
		UserMentioned,
	}
}

// Contract returns ownership metadata for a domain event type.
func Contract(eventType string) (ContractInfo, bool) {
	c, ok := contracts[eventType]
	return c, ok
}

// NewEnvelope builds a domain event envelope using the caller-owned timestamp.
func NewEnvelope(aggregateType string, aggregateID int64, eventType string, occurredAt time.Time, payload any) Envelope {
	return Envelope{
		EventID:       uuid.NewString(),
		EventType:     eventType,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		OccurredAt:    occurredAt,
		Payload:       payload,
	}
}

// MarkProcessed records consumer idempotency. Duplicate rows mean success.
func MarkProcessed(ctx context.Context, db database.DBTX, eventID, consumerName string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO processed_events (event_id, consumer_name, processed_at) VALUES (?, ?, ?)`,
		eventID,
		consumerName,
		time.Now().UTC(),
	)
	if err == nil || isDuplicate(err) {
		return nil
	}
	return fmt.Errorf("mark processed event: %w", err)
}

// IsProcessed reports whether a consumer has already handled an event.
func IsProcessed(ctx context.Context, db database.DBTX, eventID, consumerName string) (bool, error) {
	var count int
	err := db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM processed_events WHERE event_id = ? AND consumer_name = ?`,
		eventID,
		consumerName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check processed event: %w", err)
	}
	return count > 0, nil
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
