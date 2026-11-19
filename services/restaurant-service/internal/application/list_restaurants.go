package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

const (
	DefaultRestaurantsPageSize = 20
	MaxRestaurantsPageSize     = 100
	// MaxRestaurantsPage limita o offset (page * page_size): skip no Mongo e' O(offset),
	// entao paginas muito altas viram varredura cara sem utilidade para o cliente.
	MaxRestaurantsPage = 1000
)

// ListRestaurantsInput usa zero como "nao informado": o caso de uso aplica os defaults.
type ListRestaurantsInput struct {
	Page     int
	PageSize int
}

type ListRestaurantsOutput struct {
	Restaurants []entities.Restaurant
	Page        int
	PageSize    int
	HasMore     bool
}

type ListRestaurantsUseCase struct {
	restaurants ports.RestaurantRepository
}

func NewListRestaurantsUseCase(restaurants ports.RestaurantRepository) *ListRestaurantsUseCase {
	return &ListRestaurantsUseCase{restaurants: restaurants}
}

func (uc *ListRestaurantsUseCase) Execute(ctx context.Context, input ListRestaurantsInput) (ListRestaurantsOutput, error) {
	page, pageSize := input.Page, input.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = DefaultRestaurantsPageSize
	}
	if page < 1 || page > MaxRestaurantsPage {
		return ListRestaurantsOutput{}, apperrors.InvalidArgument("page must be between 1 and 1000", nil)
	}
	if pageSize < 1 || pageSize > MaxRestaurantsPageSize {
		return ListRestaurantsOutput{}, apperrors.InvalidArgument("page_size must be between 1 and 100", nil)
	}

	// Pede um registro a mais para saber se existe proxima pagina sem um
	// CountDocuments extra (uma consulta a menos por request).
	restaurants, err := uc.restaurants.List(ctx, ports.ListRestaurantsQuery{
		Offset: (page - 1) * pageSize,
		Limit:  pageSize + 1,
	})
	if err != nil {
		return ListRestaurantsOutput{}, apperrors.Internal("failed to list restaurants", err)
	}

	hasMore := len(restaurants) > pageSize
	if hasMore {
		restaurants = restaurants[:pageSize]
	}

	return ListRestaurantsOutput{
		Restaurants: restaurants,
		Page:        page,
		PageSize:    pageSize,
		HasMore:     hasMore,
	}, nil
}
