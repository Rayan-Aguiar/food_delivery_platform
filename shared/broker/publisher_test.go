package broker

import (
	"context"
	"errors"
	"strings"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type publisherMock struct {
	called    bool
	exchange  string
	routing   string
	msg       amqp.Publishing
	returnErr error
}

func (m *publisherMock) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	m.called = true
	m.exchange = exchange
	m.routing = key
	m.msg = msg
	return m.returnErr
}

func TestPublishJSON_Success(t *testing.T) {
	m := &publisherMock{}
	err := publishJSON(context.Background(), m, "order.exchange", "order.created", map[string]string{"a": "b"}, amqp.Table{"x": "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.called {
		t.Fatal("expected publisher to be called")
	}
	if m.exchange != "order.exchange" || m.routing != "order.created" {
		t.Fatalf("unexpected route: %s %s", m.exchange, m.routing)
	}
	if m.msg.ContentType != "application/json" {
		t.Fatalf("unexpected content type: %s", m.msg.ContentType)
	}
	if m.msg.DeliveryMode != amqp.Persistent {
		t.Fatalf("unexpected delivery mode: %d", m.msg.DeliveryMode)
	}
}

func TestPublishJSON_MarshalError(t *testing.T) {
	m := &publisherMock{}
	err := publishJSON(context.Background(), m, "x", "y", map[string]any{"bad": func() {}}, nil)
	if err == nil {
		t.Fatal("expected marshal error")
	}
	if !strings.Contains(err.Error(), "marshal msg") {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.called {
		t.Fatal("publisher should not be called on marshal error")
	}
}

func TestPublishJSON_PublishError(t *testing.T) {
	m := &publisherMock{returnErr: errors.New("publish failed")}
	err := publishJSON(context.Background(), m, "x", "y", map[string]string{"ok": "1"}, nil)
	if err == nil {
		t.Fatal("expected publish error")
	}
	if !strings.Contains(err.Error(), "publish") {
		t.Fatalf("unexpected error: %v", err)
	}
}
