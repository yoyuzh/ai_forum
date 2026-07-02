package event

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestNewEnvelopeUsesRequiredShape(t *testing.T) {
	occurredAt := time.Date(2026, 7, 2, 1, 2, 3, 0, time.UTC)

	env := NewEnvelope("post", 1001, PostCreated, occurredAt, map[string]any{"post_id": int64(1001)})

	if env.EventID == "" {
		t.Fatal("event id is empty")
	}
	if env.EventType != "post.created" || env.AggregateType != "post" || env.AggregateID != 1001 {
		t.Fatalf("envelope = %#v", env)
	}
	if !env.OccurredAt.Equal(occurredAt) {
		t.Fatalf("occurredAt = %s, want %s", env.OccurredAt, occurredAt)
	}
	payload := env.Payload.(map[string]any)
	if payload["post_id"] != int64(1001) {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestAllEventContractsHaveOwners(t *testing.T) {
	for _, eventType := range Types() {
		contract, ok := Contract(eventType)
		if !ok {
			t.Fatalf("missing contract for %s", eventType)
		}
		if contract.PublisherOwner == "" || contract.ConsumerOwner == "" {
			t.Fatalf("missing owner metadata for %s: %#v", eventType, contract)
		}
	}
}

func TestMarkProcessedTreatsDuplicateAsSuccess(t *testing.T) {
	db := &eventDBTX{execErr: &mysql.MySQLError{Number: 1062, Message: "duplicate"}}

	if err := MarkProcessed(context.Background(), db, "evt-1", "consumer-a"); err != nil {
		t.Fatalf("duplicate should be idempotent success: %v", err)
	}
}

func TestIsProcessed(t *testing.T) {
	db := &eventDBTX{getCount: 1}

	ok, err := IsProcessed(context.Background(), db, "evt-1", "consumer-a")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("event should be processed")
	}
}

type eventDBTX struct {
	execErr  error
	getCount int
}

func (d *eventDBTX) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return fakeSQLResult(1), d.execErr
}

func (d *eventDBTX) GetContext(_ context.Context, dest interface{}, _ string, _ ...interface{}) error {
	if d.getCount < 0 {
		return sql.ErrNoRows
	}
	*(dest.(*int)) = d.getCount
	return nil
}

func (d *eventDBTX) SelectContext(context.Context, interface{}, string, ...interface{}) error {
	return errors.New("unexpected select")
}

type fakeSQLResult int64

func (r fakeSQLResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeSQLResult) RowsAffected() (int64, error) { return int64(r), nil }
