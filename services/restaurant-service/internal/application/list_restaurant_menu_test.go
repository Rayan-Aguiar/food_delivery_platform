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

func TestListRestaurantMenuUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	validInput := application.ListRestaurantMenuInput{RestaurantID: "rest-1"}
	restaurantFound := &fakeRestaurantRepo{
		getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
			return mustRestaurant(), nil
		},
	}

	tests := []struct {
		name        string
		input       application.ListRestaurantMenuInput
		restaurants *fakeRestaurantRepo
		menu        *fakeMenuRepo
		wantCount   int
		wantErr     bool
		wantCode    string
	}{
		{
			name:        "success returns menu items",
			input:       validInput,
			restaurants: restaurantFound,
			menu: &fakeMenuRepo{
				listByRestaurantIDFn: func(_ context.Context, restaurantID string) ([]entities.MenuItem, error) {
					return []entities.MenuItem{mustMenuItem()}, nil
				},
			},
			wantCount: 1,
		},
		{
			name:        "empty id",
			input:       application.ListRestaurantMenuInput{RestaurantID: ""},
			restaurants: &fakeRestaurantRepo{},
			menu:        &fakeMenuRepo{},
			wantErr:     true,
			wantCode:    apperrors.CodeInvalidArgument,
		},
		{
			name:  "restaurant not found",
			input: validInput,
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
			name:        "menu repository error",
			input:       validInput,
			restaurants: restaurantFound,
			menu: &fakeMenuRepo{
				listByRestaurantIDFn: func(_ context.Context, restaurantID string) ([]entities.MenuItem, error) {
					return nil, errors.New("db timeout")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewListRestaurantMenuUseCase(tc.restaurants, tc.menu)

			got, err := uc.Execute(ctx, tc.input)

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
			if len(got) != tc.wantCount {
				t.Errorf("expected %d menu items, got %d", tc.wantCount, len(got))
			}
		})
	}
}
