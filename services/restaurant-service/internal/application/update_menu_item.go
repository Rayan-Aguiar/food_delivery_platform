package application

import (
	"context"
	"errors"

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
}

type UpdateMenuItemUseCase struct {
	menu ports.MenuRepository
}

func NewUpdateMenuItemUseCase(menu ports.MenuRepository) *UpdateMenuItemUseCase {
	return &UpdateMenuItemUseCase{menu: menu}
}

func (uc *UpdateMenuItemUseCase) Execute(ctx context.Context, input UpdateMenuItemInput) (entities.MenuItem, error) {
	if input.RestaurantID == "" || input.ItemID == "" {
		return entities.MenuItem{}, apperrors.InvalidArgument("restaurant id and item id are required", nil)
	}

	existing, err := uc.menu.GetByID(ctx, input.RestaurantID, input.ItemID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrMenuItemNotFound) {
			return entities.MenuItem{}, apperrors.NotFound("menu item not found", nil)
		}
		return entities.MenuItem{}, apperrors.Internal("failed to get menu item", err)
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

	return updated, nil
}
