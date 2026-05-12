package contracts

type EventMeta struct {
	EventID        string `json:"event_id"`
	EventType      string `json:"event_type"`
	EventVersion   int    `json:"event_version"`
	OccurredAt     string `json:"occurred_at"`
	Producer       string `json:"producer"`
	CorrelationID  string `json:"correlation_id"`
	CausationID    string `json:"causation_id,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	Traceparent    string `json:"traceparent,omitempty"`
}

type Event[T any] struct {
	Meta    EventMeta `json:"meta"`
	Payload T         `json:"payload"`
}
