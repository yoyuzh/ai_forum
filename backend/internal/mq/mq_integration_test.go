//go:build integration

// Integration tests for the mq package, run against a live RabbitMQ
// container (docker-compose up -d rabbitmq). Build tag `integration` keeps
// these out of the default `go test ./...` run.
//
// Run with:
//
//	RABBITMQ_URL=amqp://guest:guest@127.0.0.1:5672/ \
//	go test -tags=integration ./internal/mq/...
package mq

import (
	"context"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/config"
)

// mqCfgFromEnv builds the RabbitMQ config the same way the loader does, from
// the same env vars. Defaults match docker-compose so `docker compose up -d`
// + `go test -tags=integration` works out of the box.
func mqCfgFromEnv() config.RabbitMQConfig {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@127.0.0.1:5672/"
	}
	return config.RabbitMQConfig{URL: url}
}

// TestRabbitMQPublishConsume verifies NewRabbitMQ produces a connection that
// can declare a temp queue, publish a message, and consume it back (spec:
// rabbitmq-client, "Publish and consume round-trip").
func TestRabbitMQPublishConsume(t *testing.T) {
	// Arrange
	cfg := mqCfgFromEnv()
	conn, err := NewRabbitMQ(cfg)
	require.NoError(t, err, "NewRabbitMQ must connect to the live RabbitMQ container")
	t.Cleanup(func() { _ = conn.Close() })

	ch, err := conn.Channel()
	require.NoError(t, err, "Channel must open on a healthy connection")
	t.Cleanup(func() { _ = ch.Close() })

	queueName := "test-mq-p2-roundtrip"
	q, err := ch.QueueDeclare(
		queueName,
		false, // durable
		true,  // autoDelete
		true,  // exclusive
		false, // noWait
		nil,   // args
	)
	require.NoError(t, err, "QueueDeclare must succeed on the live broker")

	body := []byte("p2-mq-roundtrip")

	// Act — publish
	pubCtx, pubCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer pubCancel()
	require.NoError(t,
		ch.PublishWithContext(pubCtx, "", q.Name, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		}),
		"Publish must succeed on the declared queue")

	// Act — consume
	deliveries, err := ch.Consume(
		q.Name,             // queue
		"test-mq-consumer", // consumer
		true,               // autoAck
		true,               // exclusive
		false,              // noLocal
		false,              // noWait
		nil,                // args
	)
	require.NoError(t, err, "Consume must succeed on the declared queue")

	// Assert — receive within a 3s window
	recvCtx, recvCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer recvCancel()
	select {
	case d, ok := <-deliveries:
		require.True(t, ok, "delivery channel must not be closed before the message arrives")
		assert.Equal(t, body, d.Body, "consumed body must match the published body")
	case <-recvCtx.Done():
		t.Fatal("timed out waiting for the published message to be consumed")
	}
}

func TestTopologyRoutesPostCreatedToRequiredQueues(t *testing.T) {
	cfg := mqCfgFromEnv()
	conn, err := NewRabbitMQ(cfg)
	require.NoError(t, err, "NewRabbitMQ must connect to the live RabbitMQ container")
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, DeclareTopology(conn))

	ch, err := conn.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { _ = ch.Close() })

	for _, queue := range []string{QueuePostTagging, QueueSearchIndex, QueueAuditLog} {
		_, err := ch.QueuePurge(queue, false)
		require.NoError(t, err)
	}

	pub := NewPublisher(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, pub.Publish(ctx, ExchangeForumEvents, "post.created", []byte(`{"eventType":"post.created"}`)))

	for _, queue := range []string{QueuePostTagging, QueueSearchIndex, QueueAuditLog} {
		msg, ok, err := ch.Get(queue, true)
		require.NoError(t, err)
		require.True(t, ok, "queue %s should receive post.created", queue)
		assert.Equal(t, "post.created", msg.RoutingKey)
	}
}

func TestPublisherReconnectsAfterConnectionClose(t *testing.T) {
	cfg := mqCfgFromEnv()
	conn, err := NewRabbitMQ(cfg)
	require.NoError(t, err, "NewRabbitMQ must connect to the live RabbitMQ container")
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, DeclareTopology(conn))

	ch, err := conn.Channel()
	require.NoError(t, err)
	_, err = ch.QueuePurge(QueueAuditLog, false)
	require.NoError(t, err)
	require.NoError(t, ch.Close())

	pub := NewPublisher(conn)
	require.NoError(t, conn.Conn.Close(), "test should simulate a dropped connection")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, pub.Publish(ctx, ExchangeForumEvents, "post.created", []byte(`{"eventType":"post.created"}`)))

	ch, err = conn.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { _ = ch.Close() })
	msg, ok, err := ch.Get(QueueAuditLog, true)
	require.NoError(t, err)
	require.True(t, ok, "publish after reconnect should reach q.audit.log")
	assert.Equal(t, "post.created", msg.RoutingKey)
}

// TestNewRabbitMQDialError verifies that a malformed URL produces a wrapped
// dial error rather than a panic (spec: rabbitmq-client, construction).
func TestNewRabbitMQDialError(t *testing.T) {
	// Arrange — an unreachable URL with an invalid host.
	cfg := config.RabbitMQConfig{URL: "amqp://guest:guest@127.0.0.1:1/"}

	// Act
	conn, err := NewRabbitMQ(cfg)

	// Assert
	require.Error(t, err, "dialing an unreachable broker must return an error")
	assert.Nil(t, conn, "Connection must be nil on dial failure")
	assert.Contains(t, err.Error(), "rabbitmq dial",
		"error must be wrapped with the rabbitmq dial prefix")
}
