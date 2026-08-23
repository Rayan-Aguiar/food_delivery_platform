package application_test

import (
	"context"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

// ── Restaurant Repository ─────────────────────────────────────────────────────

type fakeRestaurantRepo struct {
	listFn    func(ctx context.Context) ([]entities.Restaurant, error)
	getByIDFn func(ctx context.Context, id string) (entities.Restaurant, error)
	saveFn    func(ctx context.Context, restaurant entities.Restaurant) error
}

func (f *fakeRestaurantRepo) List(ctx context.Context) ([]entities.Restaurant, error) {
	if f.listFn != nil {
		return f.listFn(ctx)
	}
	return nil, nil
}

func (f *fakeRestaurantRepo) GetByID(ctx context.Context, id string) (entities.Restaurant, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return entities.Restaurant{}, nil
}

func (f *fakeRestaurantRepo) Save(ctx context.Context, restaurant entities.Restaurant) error {
	if f.saveFn != nil {
		return f.saveFn(ctx, restaurant)
	}
	return nil
}

// ── Menu Repository ───────────────────────────────────────────────────────────

type fakeMenuRepo struct {
	listByRestaurantIDFn func(ctx context.Context, restaurantID string) ([]entities.MenuItem, error)
	getByIDFn            func(ctx context.Context, restaurantID, itemID string) (entities.MenuItem, error)
	saveFn               func(ctx context.Context, item entities.MenuItem) error
	updateFn             func(ctx context.Context, item entities.MenuItem) error
}

func (f *fakeMenuRepo) ListByRestaurantID(ctx context.Context, restaurantID string) ([]entities.MenuItem, error) {
	if f.listByRestaurantIDFn != nil {
		return f.listByRestaurantIDFn(ctx, restaurantID)
	}
	return nil, nil
}

func (f *fakeMenuRepo) GetByID(ctx context.Context, restaurantID, itemID string) (entities.MenuItem, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, restaurantID, itemID)
	}
	return entities.MenuItem{}, nil
}

func (f *fakeMenuRepo) Save(ctx context.Context, item entities.MenuItem) error {
	if f.saveFn != nil {
		return f.saveFn(ctx, item)
	}
	return nil
}

func (f *fakeMenuRepo) Update(ctx context.Context, item entities.MenuItem) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, item)
	}
	return nil
}

func mustRestaurant() entities.Restaurant {
	r, err := entities.NewRestaurant("rest-1", "Pizza Place", "Main St 123", "+5511999999999", entities.RestaurantStatusActive, 5.0)
	if err != nil {
		panic(err)
	}
	return r
}

func mustMenuItem() entities.MenuItem {
	m, err := entities.NewMenuItem("item-1", "rest-1", "Margherita", 39.9, entities.MenuCategoryPizzaria, true)
	if err != nil {
		panic(err)
	}
	return m
}
