package ports

import "context"

type CatalogEventPublisher interface {
	PublishMenuUpdated(ctx context.Context, in MenuUpdatedEvent) error
	PublishAvailabilityChanged(ctx context.Context, in AvailabilityChangedEvent) error
}

type MenuUpdatedEvent struct {
	RestaurantID string
	ItemID       string
	Name         string
	Price        float64
	Category     string

	CorrelationID  string
	CausationID    string
	Traceparent    string
	IdempotencyKey string
}

type AvailabilityChangedEvent struct {
	RestaurantID string
	ItemID       string
	Available    bool

	CorrelationID  string
	CausationID    string
	Traceparent    string
	IdempotencyKey string
}
