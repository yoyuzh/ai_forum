// Package outbox appends durable domain events for later publishing.
package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/event"
	"ai-forum/backend/internal/mq"
)

// Event is the domain event shape persisted into outbox_events.
type Event struct {
	EventID       string
	EventType     string
	AggregateType string
	AggregateID   int64
	Payload       any
}

// Append inserts one PENDING outbox event on the caller-owned transaction.
func Append(ctx context.Context, tx database.DBTX, event Event) error {
	if event.EventID == "" {
		event.EventID = uuid.NewString()
	}
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbox_events
			(event_id, event_type, aggregate_type, aggregate_id, payload, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.EventID,
		event.EventType,
		event.AggregateType,
		event.AggregateID,
		string(payload),
		"PENDING",
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("append outbox event: %w", err)
	}
	return nil
}

// Record is an outbox_events row selected for publishing.
type Record struct {
	ID            int64           `db:"id"`
	EventID       string          `db:"event_id"`
	EventType     string          `db:"event_type"`
	AggregateType string          `db:"aggregate_type"`
	AggregateID   int64           `db:"aggregate_id"`
	Payload       json.RawMessage `db:"payload"`
	CreatedAt     time.Time       `db:"created_at"`
	RetryCount    int             `db:"retry_count"`
}

// MessagePublisher is the RabbitMQ publisher surface needed by outbox.
type MessagePublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

// Options controls publisher scan behavior.
type Options struct {
	BatchSize    int
	MaxRetries   int
	ScanInterval time.Duration
}

// Publisher scans PENDING outbox rows and publishes them.
type Publisher struct {
	db      database.DBTX
	pub     MessagePublisher
	options Options
	done    chan struct{}
	once    sync.Once
}

// NewPublisher constructs an outbox publisher with conservative defaults.
func NewPublisher(db database.DBTX, pub MessagePublisher, options Options) *Publisher {
	if options.BatchSize <= 0 {
		options.BatchSize = 100
	}
	if options.MaxRetries <= 0 {
		options.MaxRetries = 3
	}
	if options.ScanInterval <= 0 {
		options.ScanInterval = time.Second
	}
	return &Publisher{db: db, pub: pub, options: options, done: make(chan struct{})}
}

// Start scans until the context is canceled or Stop is called.
func (p *Publisher) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.options.ScanInterval)
	defer ticker.Stop()
	defer p.once.Do(func() { close(p.done) })

	for {
		if err := p.ProcessOnce(ctx); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// Stop waits for the in-flight publish pass to finish.
func (p *Publisher) Stop(ctx context.Context) error {
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ProcessOnce publishes one batch of pending events.
func (p *Publisher) ProcessOnce(ctx context.Context) error {
	var records []Record
	if err := p.db.SelectContext(ctx, &records, `
		SELECT id, event_id, event_type, aggregate_type, aggregate_id, payload, created_at, retry_count
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at, id
		LIMIT ?`, p.options.BatchSize); err != nil {
		return fmt.Errorf("select pending outbox events: %w", err)
	}

	for _, record := range records {
		body, err := json.Marshal(event.Envelope{
			EventID:       record.EventID,
			EventType:     record.EventType,
			AggregateType: record.AggregateType,
			AggregateID:   record.AggregateID,
			OccurredAt:    record.CreatedAt,
			Payload:       json.RawMessage(record.Payload),
		})
		if err != nil {
			return fmt.Errorf("marshal event envelope: %w", err)
		}
		if err := p.pub.Publish(ctx, exchangeFor(record.EventType), record.EventType, body); err != nil {
			if markErr := p.markFailure(ctx, record); markErr != nil {
				return errors.Join(err, markErr)
			}
			continue
		}
		if err := p.markPublished(ctx, record.ID); err != nil {
			return err
		}
	}
	return nil
}

func exchangeFor(eventType string) string {
	switch eventType {
	case "post.tagged", "ai.reply.completed", "ai.reply.failed":
		return mq.ExchangeAIEvents
	default:
		return mq.ExchangeForumEvents
	}
}

func (p *Publisher) markPublished(ctx context.Context, id int64) error {
	_, err := p.db.ExecContext(ctx,
		`UPDATE outbox_events SET status = 'PUBLISHED', published_at = ? WHERE id = ?`,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		return fmt.Errorf("mark outbox published: %w", err)
	}
	return nil
}

func (p *Publisher) markFailure(ctx context.Context, record Record) error {
	if record.RetryCount+1 >= p.options.MaxRetries {
		_, err := p.db.ExecContext(ctx,
			`UPDATE outbox_events SET status = 'FAILED', retry_count = ? WHERE id = ?`,
			p.options.MaxRetries,
			record.ID,
		)
		if err != nil {
			return fmt.Errorf("mark outbox failed: %w", err)
		}
		return nil
	}
	_, err := p.db.ExecContext(ctx, `UPDATE outbox_events SET retry_count = retry_count + 1 WHERE id = ?`, record.ID)
	if err != nil {
		return fmt.Errorf("increment outbox retry: %w", err)
	}
	return nil
}
