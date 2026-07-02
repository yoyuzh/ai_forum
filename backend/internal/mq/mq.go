// Package mq provides the RabbitMQ connection and channel infrastructure for
// domain event publishing and consuming (architecture §3.3). P2 only
// constructs the connection and proves a round-trip; reconnect/retry logic and
// exchange/queue topology are deferred to P5.
package mq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"ai-forum/backend/internal/config"
)

const (
	ExchangeForumEvents        = "forum.events"
	ExchangeAIEvents           = "ai.events"
	ExchangeNotificationEvents = "notification.events"
	ExchangeDead               = "dead.exchange"

	QueuePostTagging  = "q.post.tagging"
	QueueAIDecision   = "q.ai.decision"
	QueueSearchIndex  = "q.search.index"
	QueueNotification = "q.notification"
	QueueAuditLog     = "q.audit.log"
	QueueDead         = "q.dead"
)

// Exchange declares a durable RabbitMQ exchange.
type Exchange struct {
	Name string
	Kind string
}

// Queue declares a durable RabbitMQ queue.
type Queue struct {
	Name string
}

// Binding declares a queue binding.
type Binding struct {
	Exchange   string
	Queue      string
	RoutingKey string
}

// TopologySpec is the P5 RabbitMQ topology contract.
type TopologySpec struct {
	Exchanges []Exchange
	Queues    []Queue
	Bindings  []Binding
}

// Topology returns the durable RabbitMQ topology for domain events.
func Topology() TopologySpec {
	return TopologySpec{
		Exchanges: []Exchange{
			{Name: ExchangeForumEvents, Kind: "topic"},
			{Name: ExchangeAIEvents, Kind: "topic"},
			{Name: ExchangeNotificationEvents, Kind: "topic"},
			{Name: ExchangeDead, Kind: "direct"},
		},
		Queues: []Queue{
			{Name: QueuePostTagging},
			{Name: QueueAIDecision},
			{Name: QueueSearchIndex},
			{Name: QueueNotification},
			{Name: QueueAuditLog},
			{Name: QueueDead},
		},
		Bindings: []Binding{
			{Exchange: ExchangeForumEvents, Queue: QueuePostTagging, RoutingKey: "post.created"},
			{Exchange: ExchangeForumEvents, Queue: QueueSearchIndex, RoutingKey: "post.*"},
			{Exchange: ExchangeForumEvents, Queue: QueueAuditLog, RoutingKey: "post.*"},
			{Exchange: ExchangeForumEvents, Queue: QueueNotification, RoutingKey: "comment.created"},
			{Exchange: ExchangeForumEvents, Queue: QueueNotification, RoutingKey: "user.mentioned"},
			{Exchange: ExchangeAIEvents, Queue: QueueAIDecision, RoutingKey: "post.tagged"},
			{Exchange: ExchangeAIEvents, Queue: QueueNotification, RoutingKey: "ai.reply.*"},
			{Exchange: ExchangeAIEvents, Queue: QueueSearchIndex, RoutingKey: "ai.reply.completed"},
			{Exchange: ExchangeDead, Queue: QueueDead, RoutingKey: "#"},
		},
	}
}

// Exchange finds an exchange by name.
func (t TopologySpec) Exchange(name string) (Exchange, bool) {
	for _, exchange := range t.Exchanges {
		if exchange.Name == name {
			return exchange, true
		}
	}
	return Exchange{}, false
}

// HasBinding reports whether the topology contains a binding.
func (t TopologySpec) HasBinding(want Binding) bool {
	for _, binding := range t.Bindings {
		if binding == want {
			return true
		}
	}
	return false
}

// Connection wraps an amqp091 Connection so callers do not depend on the
// driver type directly. The underlying connection is reconnect-safe in the
// sense that construction surfaces a clear dial error; automatic reconnect
// logic is wired in P5 where publishers and consumers live (design D1).
type Connection struct {
	URL  string
	Conn *amqp.Connection
	mu   sync.Mutex
}

// NewRabbitMQ dials the broker at cfg.URL and returns a wrapped Connection.
// A dial failure is wrapped so the caller sees the configured URL context in
// the error chain without leaking credentials.
func NewRabbitMQ(cfg config.RabbitMQConfig) (*Connection, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	return &Connection{URL: cfg.URL, Conn: conn}, nil
}

// Channel opens a new amqp Channel on the underlying connection. Callers own
// the returned channel and are responsible for closing it when done.
func (c *Connection) Channel() (*amqp.Channel, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Conn == nil || c.Conn.IsClosed() {
		if c.URL == "" {
			return nil, fmt.Errorf("rabbitmq channel: connection is nil")
		}
		conn, err := amqp.Dial(c.URL)
		if err != nil {
			return nil, fmt.Errorf("rabbitmq reconnect: %w", err)
		}
		c.Conn = conn
	}
	ch, err := c.Conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	return ch, nil
}

// Close closes the underlying connection. It is safe to call at shutdown.
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Conn == nil {
		return nil
	}
	if err := c.Conn.Close(); err != nil {
		return fmt.Errorf("rabbitmq close: %w", err)
	}
	return nil
}

// DeclareTopology declares exchanges, queues, and bindings idempotently.
func DeclareTopology(c *Connection) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	topology := Topology()
	for _, exchange := range topology.Exchanges {
		if err := ch.ExchangeDeclare(exchange.Name, exchange.Kind, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare exchange %s: %w", exchange.Name, err)
		}
	}
	for _, queue := range topology.Queues {
		if _, err := ch.QueueDeclare(queue.Name, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare queue %s: %w", queue.Name, err)
		}
	}
	for _, binding := range topology.Bindings {
		if err := ch.QueueBind(binding.Queue, binding.RoutingKey, binding.Exchange, false, nil); err != nil {
			return fmt.Errorf("bind queue %s: %w", binding.Queue, err)
		}
	}
	return nil
}

// Publisher publishes domain events with RabbitMQ publisher confirms.
type Publisher struct {
	conn *Connection
	mu   sync.Mutex
	ch   *amqp.Channel
}

// NewPublisher returns a confirming RabbitMQ publisher.
func NewPublisher(conn *Connection) *Publisher {
	return &Publisher{conn: conn}
}

// Publish sends a message and waits for broker ack.
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	if err := p.publish(ctx, exchange, routingKey, body); err != nil {
		p.resetChannel()
		return p.publish(ctx, exchange, routingKey, body)
	}
	return nil
}

func (p *Publisher) publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	ch, err := p.channel()
	if err != nil {
		return err
	}
	confirm, err := ch.PublishWithDeferredConfirmWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("rabbitmq publish: %w", err)
	}
	if confirm == nil {
		return nil
	}
	acked, err := confirm.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("rabbitmq publish confirm: %w", err)
	}
	if !acked {
		return fmt.Errorf("rabbitmq publish nack")
	}
	return nil
}

func (p *Publisher) channel() (*amqp.Channel, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil && !p.ch.IsClosed() {
		return p.ch, nil
	}
	ch, err := p.conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("rabbitmq confirm mode: %w", err)
	}
	p.ch = ch
	return ch, nil
}

func (p *Publisher) resetChannel() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
}
