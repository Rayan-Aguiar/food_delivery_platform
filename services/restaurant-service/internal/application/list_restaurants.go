package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

type ListRestaurantsUseCase struct {
	restaurants ports.RestaurantRepository
}

func NewListRestaurantsUseCase(restaurants ports.RestaurantRepository) *ListRestaurantsUseCase {
	return &ListRestaurantsUseCase{restaurants: restaurants}
}

func (uc *ListRestaurantsUseCase) Execute(ctx context.Context) ([]entities.Restaurant, error) {
	restaurants, err := uc.restaurants.List(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to list restaurants", err)
	}

	return restaurants, nil
}
