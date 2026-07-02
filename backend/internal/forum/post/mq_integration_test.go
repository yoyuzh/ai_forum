//go:build integration

package post

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/mq"
	"ai-forum/backend/internal/outbox"
)

func TestServiceCreatePostDoesNotPublishRabbitMQ(t *testing.T) {
	db := newPostIntegrationDB(t)
	conn, err := mq.NewRabbitMQ(config.RabbitMQConfig{URL: env("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ch.Close() })
	q, err := ch.QueueDeclare("test-p4-forum-no-publish", false, true, true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ch.QueuePurge(q.Name, false); err != nil {
		t.Fatal(err)
	}

	svc := NewService(NewSQLRepository(), outbox.Append)
	if err := database.RunInTx(context.Background(), db, func(tx *sqlx.Tx) error {
		_, err := svc.CreatePost(context.Background(), tx, CreateInput{AuthorID: 1, Title: "mq", Content: "body"})
		return err
	}); err != nil {
		t.Fatal(err)
	}
	inspected, err := ch.QueueInspect(q.Name)
	if err != nil {
		t.Fatal(err)
	}
	if inspected.Messages != 0 {
		t.Fatalf("queue depth = %d, want 0", inspected.Messages)
	}
}

func TestServiceCreatePostOutboxPublisherDeliversToRabbitMQ(t *testing.T) {
	db := newPostIntegrationDB(t)
	conn, err := mq.NewRabbitMQ(config.RabbitMQConfig{URL: env("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := mq.DeclareTopology(conn); err != nil {
		t.Fatal(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ch.Close() })
	if _, err := ch.QueuePurge(mq.QueuePostTagging, false); err != nil {
		t.Fatal(err)
	}

	svc := NewService(NewSQLRepository(), outbox.Append)
	var postID int64
	if err := database.RunInTx(context.Background(), db, func(tx *sqlx.Tx) error {
		p, err := svc.CreatePost(context.Background(), tx, CreateInput{AuthorID: 1, Title: "mq", Content: "body"})
		postID = p.ID
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var pending int
	if err := db.GetContext(context.Background(), &pending, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = ? AND status = 'PENDING'`, postID); err != nil {
		t.Fatal(err)
	}
	if pending != 1 {
		t.Fatalf("pending outbox rows = %d, want 1", pending)
	}

	publisher := outbox.NewPublisher(db, mq.NewPublisher(conn), outbox.Options{ScanInterval: time.Hour})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := publisher.ProcessOnce(ctx); err != nil {
		t.Fatal(err)
	}

	msg, ok, err := ch.Get(mq.QueuePostTagging, true)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("q.post.tagging did not receive post.created")
	}
	if msg.RoutingKey != "post.created" {
		t.Fatalf("routing key = %q, want post.created", msg.RoutingKey)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
