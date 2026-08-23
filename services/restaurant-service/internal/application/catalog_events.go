package application

import (
	"context"

	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

type noopCatalogEventPublisher struct{}

func (noopCatalogEventPublisher) PublishMenuUpdated(context.Context, ports.MenuUpdatedEvent) error {
	return nil
}

func (noopCatalogEventPublisher) PublishAvailabilityChanged(context.Context, ports.AvailabilityChangedEvent) error {
	return nil
}
