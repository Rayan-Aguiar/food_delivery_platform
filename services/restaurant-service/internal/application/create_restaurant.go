package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

type CreateRestaurantInput struct {
	Name        string
	Address     string
	PhoneNumber string
	Status      string
	DeliveryFee float64

	// IdempotencyKey, quando informado, e usado como ID do restaurante: reenviar a
	// mesma chave repete o upsert no mesmo documento em vez de criar duplicatas.
	IdempotencyKey string
}

type CreateRestaurantUseCase struct {
	restaurants ports.RestaurantRepository
	idGen       ports.IDGenerator
}

func NewCreateRestaurantUseCase(restaurants ports.RestaurantRepository, idGen ports.IDGenerator) *CreateRestaurantUseCase {
	return &CreateRestaurantUseCase{restaurants: restaurants, idGen: idGen}
}

func (uc *CreateRestaurantUseCase) Execute(ctx context.Context, input CreateRestaurantInput) (entities.Restaurant, error) {
	id := input.IdempotencyKey
	if id == "" {
		id = uc.idGen.NewID()
	}

	restaurant, err := entities.NewRestaurant(id, input.Name, input.Address, input.PhoneNumber, entities.RestaurantStatus(input.Status), input.DeliveryFee)
	if err != nil {
		return entities.Restaurant{}, apperrors.InvalidArgument(err.Error(), err)
	}

	if err := uc.restaurants.Save(ctx, restaurant); err != nil {
		return entities.Restaurant{}, apperrors.Internal("failed to save restaurant", err)
	}

	return restaurant, nil
}
