package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

type ListRestaurantMenuInput struct {
	RestaurantID string
}

type ListRestaurantMenuUseCase struct {
	restaurants ports.RestaurantRepository
	menu        ports.MenuRepository
}

func NewListRestaurantMenuUseCase(restaurants ports.RestaurantRepository, menu ports.MenuRepository) *ListRestaurantMenuUseCase {
	return &ListRestaurantMenuUseCase{restaurants: restaurants, menu: menu}
}

func (uc *ListRestaurantMenuUseCase) Execute(ctx context.Context, input ListRestaurantMenuInput) ([]entities.MenuItem, error) {
	if input.RestaurantID == "" {
		return nil, apperrors.InvalidArgument("restaurant id is required", nil)
	}

	if _, err := uc.restaurants.GetByID(ctx, input.RestaurantID); err != nil {
		return nil, mapNotFoundOrInternal(err, domainerrors.ErrRestaurantNotFound, "restaurant not found", "failed to get restaurant")
	}

	items, err := uc.menu.ListByRestaurantID(ctx, input.RestaurantID)
	if err != nil {
		return nil, apperrors.Internal("failed to list restaurant menu", err)
	}

	return items, nil
}
