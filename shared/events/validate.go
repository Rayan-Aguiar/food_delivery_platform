package events

import (
    "errors"

    "food_delivery_platform/shared/contracts"
)

func ValidateMeta(meta contracts.EventMeta) error {
    if meta.EventID == "" {
        return errors.New("event_id is required")
    }
    if meta.EventType == "" {
        return errors.New("event_type is required")
    }
    if meta.EventVersion <= 0 {
        return errors.New("event_version must be > 0")
    }
    if meta.OccurredAt == "" {
        return errors.New("occurred_at is required")
    }
    if meta.Producer == "" {
        return errors.New("producer is required")
    }
    if meta.CorrelationID == "" {
        return errors.New("correlation_id is required")
    }
    return nil
}