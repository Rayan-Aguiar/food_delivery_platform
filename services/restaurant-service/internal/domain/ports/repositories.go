package ports

import (
	"context"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

type RestaurantRepository interface {
	List(ctx context.Context) ([]entities.Restaurant, error)
	GetByID(ctx context.Context, id string) (entities.Restaurant, error)
	Save(ctx context.Context, restaurant entities.Restaurant) error
}

type MenuRepository interface {
	ListByRestaurantID(ctx context.Context, restaurantID string) ([]entities.MenuItem, error)
	GetByID(ctx context.Context, restaurantID, itemID string) (entities.MenuItem, error)
	Save(ctx context.Context, item entities.MenuItem) error
	Update(ctx context.Context, item entities.MenuItem) error
}