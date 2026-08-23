package application_test

import (
	"context"
	"errors"
	"testing"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
)

func mustAvailableItem(id string, price float64) entities.MenuItem {
	m, err := entities.NewMenuItem(id, "rest-1", "Item "+id, price, entities.MenuCategoryPizzaria, true)
	if err != nil {
		panic(err)
	}
	return m
}

func mustUnavailableItem(id string, price float64) entities.MenuItem {
	m, err := entities.NewMenuItem(id, "rest-1", "Item "+id, price, entities.MenuCategoryPizzaria, false)
	if err != nil {
		panic(err)
	}
	return m
}

func TestValidateMenuItemsUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	activeRestaurant := &fakeRestaurantRepo{
		getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
			return mustRestaurant(), nil
		},
	}

	tests := []struct {
		name        string
		input       application.ValidateMenuItemsInput
		restaurants *fakeRestaurantRepo
		menu        *fakeMenuRepo
		wantErr     bool
		wantCode    string
		check       func(t *testing.T, out application.ValidateMenuItemsOutput)
	}{
		{
			name:        "empty restaurant id",
			input:       application.ValidateMenuItemsInput{ItemIDs: []string{"item-1"}},
			restaurants: &fakeRestaurantRepo{},
			menu:        &fakeMenuRepo{},
			wantErr:     true,
			wantCode:    apperrors.CodeInvalidArgument,
		},
		{
			name:        "empty item list",
			input:       application.ValidateMenuItemsInput{RestaurantID: "rest-1"},
			restaurants: &fakeRestaurantRepo{},
			menu:        &fakeMenuRepo{},
			wantErr:     true,
			wantCode:    apperrors.CodeInvalidArgument,
		},
		{
			name:  "restaurant not found",
			input: application.ValidateMenuItemsInput{RestaurantID: "missing", ItemIDs: []string{"item-1"}},
			restaurants: &fakeRestaurantRepo{
				getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
					return entities.Restaurant{}, domainerrors.ErrRestaurantNotFound
				},
			},
			menu:     &fakeMenuRepo{},
			wantErr:  true,
			wantCode: apperrors.CodeNotFound,
		},
		{
			name:        "all items available and valid",
			input:       application.ValidateMenuItemsInput{RestaurantID: "rest-1", ItemIDs: []string{"item-1", "item-2"}},
			restaurants: activeRestaurant,
			menu: &fakeMenuRepo{
				listByRestaurantIDFn: func(_ context.Context, _ string) ([]entities.MenuItem, error) {
					return []entities.MenuItem{mustAvailableItem("item-1", 10), mustAvailableItem("item-2", 20)}, nil
				},
			},
			check: func(t *testing.T, out application.ValidateMenuItemsOutput) {
				if !out.AllValid {
					t.Fatal("expected AllValid = true")
				}
				if !out.RestaurantActive {
					t.Error("expected RestaurantActive = true")
				}
				if len(out.Items) != 2 {
					t.Fatalf("expected 2 items, got %d", len(out.Items))
				}
				for _, item := range out.Items {
					if !item.Valid || !item.Found || !item.Available {
						t.Errorf("expected item %q valid/found/available, got %+v", item.ItemID, item)
					}
				}
				if out.Items[0].Price != 10 || out.Items[1].Price != 20 {
					t.Errorf("unexpected prices: %+v", out.Items)
				}
			},
		},
		{
			name:        "unavailable item invalidates only itself",
			input:       application.ValidateMenuItemsInput{RestaurantID: "rest-1", ItemIDs: []string{"item-1", "item-2"}},
			restaurants: activeRestaurant,
			menu: &fakeMenuRepo{
				listByRestaurantIDFn: func(_ context.Context, _ string) ([]entities.MenuItem, error) {
					return []entities.MenuItem{mustAvailableItem("item-1", 10), mustUnavailableItem("item-2", 20)}, nil
				},
			},
			check: func(t *testing.T, out application.ValidateMenuItemsOutput) {
				if out.AllValid {
					t.Fatal("expected AllValid = false")
				}
				if !out.Items[0].Valid {
					t.Error("expected item-1 to be valid")
				}
				if out.Items[1].Valid {
					t.Error("expected item-2 to be invalid")
				}
				if out.Items[1].Reason != application.ValidationReasonItemUnavailable {
					t.Errorf("reason = %q, want %q", out.Items[1].Reason, application.ValidationReasonItemUnavailable)
				}
			},
		},
		{
			name:        "item not found in menu",
			input:       application.ValidateMenuItemsInput{RestaurantID: "rest-1", ItemIDs: []string{"item-x"}},
			restaurants: activeRestaurant,
			menu: &fakeMenuRepo{
				listByRestaurantIDFn: func(_ context.Context, _ string) ([]entities.MenuItem, error) {
					return []entities.MenuItem{mustAvailableItem("item-1", 10)}, nil
				},
			},
			check: func(t *testing.T, out application.ValidateMenuItemsOutput) {
				if out.AllValid {
					t.Fatal("expected AllValid = false")
				}
				if out.Items[0].Found || out.Items[0].Valid {
					t.Errorf("expected item-x not found and invalid, got %+v", out.Items[0])
				}
				if out.Items[0].Reason != application.ValidationReasonItemNotFound {
					t.Errorf("reason = %q, want %q", out.Items[0].Reason, application.ValidationReasonItemNotFound)
				}
			},
		},
		{
			name:  "inactive restaurant invalidates every item",
			input: application.ValidateMenuItemsInput{RestaurantID: "rest-1", ItemIDs: []string{"item-1"}},
			restaurants: &fakeRestaurantRepo{
				getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
					r, err := entities.NewRestaurant("rest-1", "Pizza Place", "Main St 123", "+5511999999999", entities.RestaurantStatusInactive, 5.0)
					if err != nil {
						t.Fatalf("build restaurant: %v", err)
					}
					return r, nil
				},
			},
			menu: &fakeMenuRepo{
				listByRestaurantIDFn: func(_ context.Context, _ string) ([]entities.MenuItem, error) {
					return []entities.MenuItem{mustAvailableItem("item-1", 10)}, nil
				},
			},
			check: func(t *testing.T, out application.ValidateMenuItemsOutput) {
				if out.RestaurantActive {
					t.Error("expected RestaurantActive = false")
				}
				if out.AllValid {
					t.Fatal("expected AllValid = false")
				}
				if out.Items[0].Valid {
					t.Error("expected item-1 to be invalid due to inactive restaurant")
				}
				if out.Items[0].Reason != application.ValidationReasonRestaurantInactive {
					t.Errorf("reason = %q, want %q", out.Items[0].Reason, application.ValidationReasonRestaurantInactive)
				}
			},
		},
		{
			name:        "menu repository error maps to internal",
			input:       application.ValidateMenuItemsInput{RestaurantID: "rest-1", ItemIDs: []string{"item-1"}},
			restaurants: activeRestaurant,
			menu: &fakeMenuRepo{
				listByRestaurantIDFn: func(_ context.Context, _ string) ([]entities.MenuItem, error) {
					return nil, errors.New("db unavailable")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewValidateMenuItemsUseCase(tc.restaurants, tc.menu)

			out, err := uc.Execute(ctx, tc.input)

			if tc.wantErr {
				var appErr *apperrors.AppError
				if !errors.As(err, &appErr) {
					t.Fatalf("expected *AppError, got %T: %v", err, err)
				}
				if appErr.Code != tc.wantCode {
					t.Errorf("expected code %q, got %q", tc.wantCode, appErr.Code)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tc.check(t, out)
		})
	}
}
