package events

import (
    "time"

    "github.com/google/uuid"

    "food_delivery_platform/shared/contracts"
)

func NewEnvelope[T any](
    eventType string,
    version int,
    producer string,
    correlationID string,
    causationID string,
    idempotencyKey string,
    traceparent string,
    payload T,
) contracts.Event[T] {
    if correlationID == "" {
        correlationID = uuid.NewString()
    }

    return contracts.Event[T]{
        Meta: contracts.EventMeta{
            EventID:        uuid.NewString(),
            EventType:      eventType,
            EventVersion:   version,
            OccurredAt:     time.Now().UTC().Format(time.RFC3339Nano),
            Producer:       producer,
            CorrelationID:  correlationID,
            CausationID:    causationID,
            IdempotencyKey: idempotencyKey,
            Traceparent:    traceparent,
        },
        Payload: payload,
    }
}