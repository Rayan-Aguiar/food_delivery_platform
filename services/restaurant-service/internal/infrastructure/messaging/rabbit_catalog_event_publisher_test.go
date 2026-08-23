package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
	"food_delivery_platform/shared/broker"

	amqp "github.com/rabbitmq/amqp091-go"
)

type fakeOutboxRepo struct {
	called bool
	msg    ports.OutboxMessage
	err    error
}

func (f *fakeOutboxRepo) SavePending(_ context.Context, msg ports.OutboxMessage) error {
	f.called = true
	f.msg = msg
	return f.err
}

func TestPublishMenuUpdated_Success(t *testing.T) {
	p := NewRabbitCatalogEventPublisher(nil, "restaurant.exchange", "restaurant-service", nil, broker.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})

	called := 0
	p.publishFn = func(_ context.Context, _ *amqp.Channel, exchange, routingKey string, msg any, headers amqp.Table) error {
		called++
		if exchange != "restaurant.exchange" {
			t.Fatalf("exchange = %q, want %q", exchange, "restaurant.exchange")
		}
		if routingKey != "restaurant.menu.updated" {
			t.Fatalf("routing key = %q, want %q", routingKey, "restaurant.menu.updated")
		}
		b, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal event: %v", err)
		}
		if !strings.Contains(string(b), "restaurant.menu.updated.v1") {
			t.Fatalf("expected event type in payload, got: %s", string(b))
		}
		if headers["attempt"] != int32(1) {
			t.Fatalf("attempt header = %v, want 1", headers["attempt"])
		}
		if headers["correlation_id"] != "corr-1" {
			t.Fatalf("correlation_id header = %v, want corr-1", headers["correlation_id"])
		}
		return nil
	}

	err := p.PublishMenuUpdated(context.Background(), ports.MenuUpdatedEvent{
		RestaurantID:  "rest-1",
		ItemID:        "item-1",
		Name:          "Margherita",
		Price:         39.9,
		Category:      "pizzaria",
		CorrelationID: "corr-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Fatalf("publish calls = %d, want 1", called)
	}
}

func TestPublishAvailabilityChanged_RetryUntilSuccess(t *testing.T) {
	p := NewRabbitCatalogEventPublisher(nil, "restaurant.exchange", "restaurant-service", nil, broker.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})

	calls := 0
	p.publishFn = func(_ context.Context, _ *amqp.Channel, exchange, routingKey string, msg any, headers amqp.Table) error {
		calls++
		if routingKey != "restaurant.availability.changed" {
			t.Fatalf("routing key = %q, want %q", routingKey, "restaurant.availability.changed")
		}
		b, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal event: %v", err)
		}
		if !strings.Contains(string(b), "restaurant.availability.changed.v1") {
			t.Fatalf("expected event type in payload, got: %s", string(b))
		}
		if calls < 3 {
			return errors.New("temporary publish failure")
		}
		if headers["attempt"] != int32(3) {
			t.Fatalf("attempt header = %v, want 3", headers["attempt"])
		}
		return nil
	}
	p.waitFn = func(context.Context, time.Duration) error { return nil }

	err := p.PublishAvailabilityChanged(context.Background(), ports.AvailabilityChangedEvent{
		RestaurantID: "rest-1",
		ItemID:       "item-1",
		Available:    false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("publish calls = %d, want 3", calls)
	}
}

func TestPublishMenuUpdated_FailureAfterRetries(t *testing.T) {
	p := NewRabbitCatalogEventPublisher(nil, "restaurant.exchange", "restaurant-service", nil, broker.RetryPolicy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})

	calls := 0
	p.publishFn = func(context.Context, *amqp.Channel, string, string, any, amqp.Table) error {
		calls++
		return errors.New("broker down")
	}
	p.waitFn = func(context.Context, time.Duration) error { return nil }

	err := p.PublishMenuUpdated(context.Background(), ports.MenuUpdatedEvent{
		RestaurantID: "rest-1",
		ItemID:       "item-1",
		Name:         "Margherita",
		Price:        39.9,
		Category:     "pizzaria",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 2 {
		t.Fatalf("publish calls = %d, want 2", calls)
	}
}

func TestPublishMenuUpdated_FallbackToOutboxOnPermanentFailure(t *testing.T) {
	outbox := &fakeOutboxRepo{}
	p := NewRabbitCatalogEventPublisher(nil, "restaurant.exchange", "restaurant-service", nil, broker.RetryPolicy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, outbox)

	p.publishFn = func(context.Context, *amqp.Channel, string, string, any, amqp.Table) error {
		return errors.New("broker down")
	}
	p.waitFn = func(context.Context, time.Duration) error { return nil }

	err := p.PublishMenuUpdated(context.Background(), ports.MenuUpdatedEvent{
		RestaurantID:  "rest-1",
		ItemID:        "item-1",
		Name:          "Margherita",
		Price:         39.9,
		Category:      "pizzaria",
		CorrelationID: "corr-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !outbox.called {
		t.Fatal("expected outbox fallback to be called")
	}
	if outbox.msg.Exchange != "restaurant.exchange" {
		t.Fatalf("outbox exchange = %q, want %q", outbox.msg.Exchange, "restaurant.exchange")
	}
	if outbox.msg.RoutingKey != "restaurant.menu.updated" {
		t.Fatalf("outbox routing key = %q, want %q", outbox.msg.RoutingKey, "restaurant.menu.updated")
	}
	if outbox.msg.Headers["correlation_id"] != "corr-1" {
		t.Fatalf("outbox correlation_id = %q, want %q", outbox.msg.Headers["correlation_id"], "corr-1")
	}
}
