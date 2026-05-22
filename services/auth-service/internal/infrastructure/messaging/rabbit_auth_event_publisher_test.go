package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"food_delivery_platform/services/auth-service/internal/domain/ports"
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

func TestPublishUserRegistered_Success(t *testing.T) {
	p := NewRabbitAuthEventPublisher(nil, "auth.exchange", "auth-service", nil, broker.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})

	called := 0
	p.publishFn = func(_ context.Context, _ *amqp.Channel, exchange, routingKey string, msg any, headers amqp.Table) error {
		called++
		if exchange != "auth.exchange" {
			t.Fatalf("exchange = %q, want %q", exchange, "auth.exchange")
		}
		if routingKey != "user.auth.registered" {
			t.Fatalf("routing key = %q, want %q", routingKey, "user.auth.registered")
		}
		b, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal event: %v", err)
		}
		if !strings.Contains(string(b), "user.auth.registered.v1") {
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

	err := p.PublishUserRegistered(context.Background(), ports.UserRegisteredEvent{
		UserID:        "user-1",
		Email:         "user@example.com",
		RegisteredAt:  time.Now().UTC().Format(time.RFC3339Nano),
		CorrelationID: "corr-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Fatalf("publish calls = %d, want 1", called)
	}
}

func TestPublishLoginSucceeded_RetryUntilSuccess(t *testing.T) {
	p := NewRabbitAuthEventPublisher(nil, "auth.exchange", "auth-service", nil, broker.RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})

	calls := 0
	p.publishFn = func(_ context.Context, _ *amqp.Channel, exchange, routingKey string, msg any, headers amqp.Table) error {
		calls++
		if exchange != "auth.exchange" {
			t.Fatalf("exchange = %q, want %q", exchange, "auth.exchange")
		}
		if routingKey != "auth.login.succeeded" {
			t.Fatalf("routing key = %q, want %q", routingKey, "auth.login.succeeded")
		}
		b, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal event: %v", err)
		}
		if !strings.Contains(string(b), "auth.login.succeeded.v1") {
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

	err := p.PublishLoginSucceeded(context.Background(), ports.LoginSucceededEvent{
		UserID:   "user-1",
		LoggedAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("publish calls = %d, want 3", calls)
	}
}

func TestPublishUserRegistered_FailureAfterRetries(t *testing.T) {
	p := NewRabbitAuthEventPublisher(nil, "auth.exchange", "auth-service", nil, broker.RetryPolicy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})

	calls := 0
	p.publishFn = func(context.Context, *amqp.Channel, string, string, any, amqp.Table) error {
		calls++
		return errors.New("broker down")
	}
	p.waitFn = func(context.Context, time.Duration) error { return nil }

	err := p.PublishUserRegistered(context.Background(), ports.UserRegisteredEvent{
		UserID:       "user-1",
		Email:        "user@example.com",
		RegisteredAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 2 {
		t.Fatalf("publish calls = %d, want 2", calls)
	}
}

func TestPublishUserRegistered_FallbackToOutboxOnPermanentFailure(t *testing.T) {
	outbox := &fakeOutboxRepo{}
	p := NewRabbitAuthEventPublisher(nil, "auth.exchange", "auth-service", nil, broker.RetryPolicy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, outbox)

	p.publishFn = func(context.Context, *amqp.Channel, string, string, any, amqp.Table) error {
		return errors.New("broker down")
	}
	p.waitFn = func(context.Context, time.Duration) error { return nil }

	err := p.PublishUserRegistered(context.Background(), ports.UserRegisteredEvent{
		UserID:        "user-1",
		Email:         "user@example.com",
		RegisteredAt:  time.Now().UTC().Format(time.RFC3339Nano),
		CorrelationID: "corr-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !outbox.called {
		t.Fatal("expected outbox fallback to be called")
	}
	if outbox.msg.Exchange != "auth.exchange" {
		t.Fatalf("outbox exchange = %q, want %q", outbox.msg.Exchange, "auth.exchange")
	}
	if outbox.msg.RoutingKey != "user.auth.registered" {
		t.Fatalf("outbox routing key = %q, want %q", outbox.msg.RoutingKey, "user.auth.registered")
	}
	if outbox.msg.Headers["correlation_id"] != "corr-1" {
		t.Fatalf("outbox correlation_id = %q, want %q", outbox.msg.Headers["correlation_id"], "corr-1")
	}
}
