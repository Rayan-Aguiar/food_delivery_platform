package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

type UpdateMenuItemInput struct {
	RestaurantID string
	ItemID       string

	// Campos opcionais: apenas os informados (nao-nil) sao alterados.
	Name      *string
	Price     *float64
	Category  *string
	Available *bool

	CorrelationID string
	CausationID   string
	Traceparent   string
}

type UpdateMenuItemUseCase struct {
	menu   ports.MenuRepository
	events ports.CatalogEventPublisher
}

func NewUpdateMenuItemUseCase(menu ports.MenuRepository, eventPublisher ...ports.CatalogEventPublisher) *UpdateMenuItemUseCase {
	publisher := ports.CatalogEventPublisher(noopCatalogEventPublisher{})
	if len(eventPublisher) > 0 && eventPublisher[0] != nil {
		publisher = eventPublisher[0]
	}

	return &UpdateMenuItemUseCase{menu: menu, events: publisher}
}

func (uc *UpdateMenuItemUseCase) Execute(ctx context.Context, input UpdateMenuItemInput) (entities.MenuItem, error) {
	if input.RestaurantID == "" || input.ItemID == "" {
		return entities.MenuItem{}, apperrors.InvalidArgument("restaurant id and item id are required", nil)
	}

	existing, err := uc.menu.GetByID(ctx, input.RestaurantID, input.ItemID)
	if err != nil {
		return entities.MenuItem{}, mapNotFoundOrInternal(err, domainerrors.ErrMenuItemNotFound, "menu item not found", "failed to get menu item")
	}

	name := existing.Name
	if input.Name != nil {
		name = *input.Name
	}
	price := existing.Price
	if input.Price != nil {
		price = *input.Price
	}
	category := existing.Category
	if input.Category != nil {
		category = entities.MenuCategory(*input.Category)
	}
	available := existing.Available
	if input.Available != nil {
		available = *input.Available
	}

	updated, err := entities.NewMenuItem(existing.ID, existing.RestaurantID, name, price, category, available)
	if err != nil {
		return entities.MenuItem{}, apperrors.InvalidArgument(err.Error(), err)
	}

	if err := uc.menu.Update(ctx, updated); err != nil {
		return entities.MenuItem{}, apperrors.Internal("failed to update menu item", err)
	}

	// Best effort: os eventos sao publicados so quando o campo relevante realmente
	// mudou, e uma falha de publicacao nao desfaz a atualizacao ja persistida.
	if existing.Available != updated.Available {
		_ = uc.events.PublishAvailabilityChanged(ctx, ports.AvailabilityChangedEvent{
			RestaurantID:   updated.RestaurantID,
			ItemID:         updated.ID,
			Available:      updated.Available,
			CorrelationID:  input.CorrelationID,
			CausationID:    input.CausationID,
			Traceparent:    input.Traceparent,
			IdempotencyKey: updated.ID,
		})
	}
	if existing.Name != updated.Name || existing.Price != updated.Price || existing.Category != updated.Category {
		_ = uc.events.PublishMenuUpdated(ctx, ports.MenuUpdatedEvent{
			RestaurantID:   updated.RestaurantID,
			ItemID:         updated.ID,
			Name:           updated.Name,
			Price:          updated.Price,
			Category:       string(updated.Category),
			CorrelationID:  input.CorrelationID,
			CausationID:    input.CausationID,
			Traceparent:    input.Traceparent,
			IdempotencyKey: updated.ID,
		})
	}

	return updated, nil
}
