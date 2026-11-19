package ports

import (
	"context"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

// ListRestaurantsQuery pagina por offset/limit com ordem estavel (por id).
type ListRestaurantsQuery struct {
	Offset int
	Limit  int
}

type RestaurantRepository interface {
	List(ctx context.Context, query ListRestaurantsQuery) ([]entities.Restaurant, error)
	// GetByID retorna domainerrors.ErrRestaurantNotFound quando nenhum restaurante corresponde ao id.
	GetByID(ctx context.Context, id string) (entities.Restaurant, error)
	Save(ctx context.Context, restaurant entities.Restaurant) error
}

type MenuRepository interface {
	ListByRestaurantID(ctx context.Context, restaurantID string) ([]entities.MenuItem, error)
	// GetByID retorna domainerrors.ErrMenuItemNotFound quando nenhum item corresponde ao id.
	GetByID(ctx context.Context, restaurantID, itemID string) (entities.MenuItem, error)
	// Save retorna domainerrors.ErrMenuItemExists quando o id ja esta em uso.
	Save(ctx context.Context, item entities.MenuItem) error
	Update(ctx context.Context, item entities.MenuItem) error
}
