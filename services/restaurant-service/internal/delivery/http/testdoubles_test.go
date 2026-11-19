package httpdelivery

import (
	"context"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

type fakeListRestaurants struct {
	fn func(ctx context.Context, input application.ListRestaurantsInput) (application.ListRestaurantsOutput, error)
}

func (f *fakeListRestaurants) Execute(ctx context.Context, input application.ListRestaurantsInput) (application.ListRestaurantsOutput, error) {
	if f.fn != nil {
		return f.fn(ctx, input)
	}
	return application.ListRestaurantsOutput{
		Restaurants: []entities.Restaurant{mustRestaurant()},
		Page:        1,
		PageSize:    application.DefaultRestaurantsPageSize,
	}, nil
}

type fakeGetRestaurantDetails struct {
	fn func(ctx context.Context, input application.GetRestaurantDetailsInput) (entities.Restaurant, error)
}

func (f *fakeGetRestaurantDetails) Execute(ctx context.Context, input application.GetRestaurantDetailsInput) (entities.Restaurant, error) {
	if f.fn != nil {
		return f.fn(ctx, input)
	}
	return mustRestaurant(), nil
}

type fakeListRestaurantMenu struct {
	fn func(ctx context.Context, input application.ListRestaurantMenuInput) ([]entities.MenuItem, error)
}

func (f *fakeListRestaurantMenu) Execute(ctx context.Context, input application.ListRestaurantMenuInput) ([]entities.MenuItem, error) {
	if f.fn != nil {
		return f.fn(ctx, input)
	}
	return []entities.MenuItem{mustMenuItem()}, nil
}

type fakeCreateRestaurant struct {
	fn func(ctx context.Context, input application.CreateRestaurantInput) (entities.Restaurant, error)
}

func (f *fakeCreateRestaurant) Execute(ctx context.Context, input application.CreateRestaurantInput) (entities.Restaurant, error) {
	if f.fn != nil {
		return f.fn(ctx, input)
	}
	return mustRestaurant(), nil
}

type fakeCreateMenuItem struct {
	fn func(ctx context.Context, input application.CreateMenuItemInput) (entities.MenuItem, error)
}

func (f *fakeCreateMenuItem) Execute(ctx context.Context, input application.CreateMenuItemInput) (entities.MenuItem, error) {
	if f.fn != nil {
		return f.fn(ctx, input)
	}
	return mustMenuItem(), nil
}

type fakeUpdateMenuItem struct {
	fn func(ctx context.Context, input application.UpdateMenuItemInput) (entities.MenuItem, error)
}

func (f *fakeUpdateMenuItem) Execute(ctx context.Context, input application.UpdateMenuItemInput) (entities.MenuItem, error) {
	if f.fn != nil {
		return f.fn(ctx, input)
	}
	return mustMenuItem(), nil
}

type fakeValidateMenuItems struct {
	fn func(ctx context.Context, input application.ValidateMenuItemsInput) (application.ValidateMenuItemsOutput, error)
}

func (f *fakeValidateMenuItems) Execute(ctx context.Context, input application.ValidateMenuItemsInput) (application.ValidateMenuItemsOutput, error) {
	if f.fn != nil {
		return f.fn(ctx, input)
	}
	return application.ValidateMenuItemsOutput{
		RestaurantID:     input.RestaurantID,
		RestaurantActive: true,
		AllValid:         true,
		Items: []application.MenuItemValidationResult{
			{ItemID: "item-1", Found: true, Available: true, Price: 39.9, Valid: true},
		},
	}, nil
}

type fakeTokenValidator struct {
	fn func(ctx context.Context, token string) (string, error)
}

func (f *fakeTokenValidator) ValidateAccessToken(ctx context.Context, token string) (string, error) {
	if f.fn != nil {
		return f.fn(ctx, token)
	}
	return "user-1", nil
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
