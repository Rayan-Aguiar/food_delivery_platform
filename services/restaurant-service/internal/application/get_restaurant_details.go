package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

type GetRestaurantDetailsInput struct {
	RestaurantID string
}

type GetRestaurantDetailsUseCase struct {
	restaurants ports.RestaurantRepository
}

func NewGetRestaurantDetailsUseCase(restaurants ports.RestaurantRepository) *GetRestaurantDetailsUseCase {
	return &GetRestaurantDetailsUseCase{restaurants: restaurants}
}

func (uc *GetRestaurantDetailsUseCase) Execute(ctx context.Context, input GetRestaurantDetailsInput) (entities.Restaurant, error) {
	if input.RestaurantID == "" {
		return entities.Restaurant{}, apperrors.InvalidArgument("restaurant id is required", nil)
	}

	restaurant, err := uc.restaurants.GetByID(ctx, input.RestaurantID)
	if err != nil {
		return entities.Restaurant{}, mapNotFoundOrInternal(err, domainerrors.ErrRestaurantNotFound, "restaurant not found", "failed to get restaurant")
	}

	return restaurant, nil
}
