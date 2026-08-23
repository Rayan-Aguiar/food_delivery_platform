package application_test

import (
	"context"
	"errors"
	"testing"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

func TestListRestaurantsUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		repo      *fakeRestaurantRepo
		wantCount int
		wantErr   bool
		wantCode  string
	}{
		{
			name: "success returns restaurants",
			repo: &fakeRestaurantRepo{
				listFn: func(_ context.Context) ([]entities.Restaurant, error) {
					return []entities.Restaurant{mustRestaurant()}, nil
				},
			},
			wantCount: 1,
		},
		{
			name: "success with empty list",
			repo: &fakeRestaurantRepo{
				listFn: func(_ context.Context) ([]entities.Restaurant, error) {
					return []entities.Restaurant{}, nil
				},
			},
			wantCount: 0,
		},
		{
			name: "repository error",
			repo: &fakeRestaurantRepo{
				listFn: func(_ context.Context) ([]entities.Restaurant, error) {
					return nil, errors.New("db unavailable")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewListRestaurantsUseCase(tc.repo)

			got, err := uc.Execute(ctx)

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
				t.Errorf("expected %d restaurants, got %d", tc.wantCount, len(got))
			}
		})
	}
}
