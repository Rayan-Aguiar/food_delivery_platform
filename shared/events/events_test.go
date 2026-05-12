package events

import (
	"testing"
	"time"

	"food_delivery_platform/shared/contracts"
)

func TestNewEnvelope_WithProvidedCorrelationID(t *testing.T) {
	e := NewEnvelope("order.created.v1", 1, "order-service", "corr-1", "cause-1", "idem-1", "tp", map[string]string{"a": "b"})
	if e.Meta.CorrelationID != "corr-1" {
		t.Fatalf("unexpected correlation id: %s", e.Meta.CorrelationID)
	}
	if e.Meta.EventID == "" || e.Meta.OccurredAt == "" {
		t.Fatal("expected generated event id and occurred at")
	}
	if _, err := time.Parse(time.RFC3339Nano, e.Meta.OccurredAt); err != nil {
		t.Fatalf("invalid occurred_at format: %v", err)
	}
}

func TestNewEnvelope_GeneratesCorrelationIDWhenEmpty(t *testing.T) {
	e := NewEnvelope("order.created.v1", 1, "order-service", "", "", "", "", map[string]string{"a": "b"})
	if e.Meta.CorrelationID == "" {
		t.Fatal("expected generated correlation id")
	}
}

func TestValidateMeta(t *testing.T) {
	valid := contracts.EventMeta{
		EventID:       "id",
		EventType:     "x",
		EventVersion:  1,
		OccurredAt:    "now",
		Producer:      "svc",
		CorrelationID: "corr",
	}
	if err := ValidateMeta(valid); err != nil {
		t.Fatalf("expected valid meta, got: %v", err)
	}

	tests := []contracts.EventMeta{
		{EventType: "x", EventVersion: 1, OccurredAt: "now", Producer: "svc", CorrelationID: "corr"},
		{EventID: "id", EventVersion: 1, OccurredAt: "now", Producer: "svc", CorrelationID: "corr"},
		{EventID: "id", EventType: "x", EventVersion: 0, OccurredAt: "now", Producer: "svc", CorrelationID: "corr"},
		{EventID: "id", EventType: "x", EventVersion: 1, Producer: "svc", CorrelationID: "corr"},
		{EventID: "id", EventType: "x", EventVersion: 1, OccurredAt: "now", CorrelationID: "corr"},
		{EventID: "id", EventType: "x", EventVersion: 1, OccurredAt: "now", Producer: "svc"},
	}

	for i, meta := range tests {
		if err := ValidateMeta(meta); err == nil {
			t.Fatalf("case %d expected error", i)
		}
	}
}

func TestEventNames_NotEmptyAndUnique(t *testing.T) {
	names := []string{
		OrderCreated,
		OrderConfirmed,
		OrderCancelled,
		OrderStatusChanged,
		PaymentApproved,
		PaymentFailed,
		PaymentRefunded,
		PaymentRefundReq,
		DeliveryRequested,
		DeliveryStarted,
		DeliveryCompleted,
		DeliveryFailed,
		UserAuthRegistered,
		NotificationSent,
		NotificationFailed,
	}

	seen := map[string]bool{}
	for _, n := range names {
		if n == "" {
			t.Fatal("event name must not be empty")
		}
		if seen[n] {
			t.Fatalf("duplicated event name: %s", n)
		}
		seen[n] = true
	}
}
