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

func TestGetRestaurantDetailsUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		input    application.GetRestaurantDetailsInput
		repo     *fakeRestaurantRepo
		wantID   string
		wantErr  bool
		wantCode string
	}{
		{
			name:  "success",
			input: application.GetRestaurantDetailsInput{RestaurantID: "rest-1"},
			repo: &fakeRestaurantRepo{
				getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
					return mustRestaurant(), nil
				},
			},
			wantID: "rest-1",
		},
		{
			name:     "empty id",
			input:    application.GetRestaurantDetailsInput{RestaurantID: ""},
			repo:     &fakeRestaurantRepo{},
			wantErr:  true,
			wantCode: apperrors.CodeInvalidArgument,
		},
		{
			name:  "restaurant not found",
			input: application.GetRestaurantDetailsInput{RestaurantID: "missing"},
			repo: &fakeRestaurantRepo{
				getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
					return entities.Restaurant{}, domainerrors.ErrRestaurantNotFound
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeNotFound,
		},
		{
			name:  "repository error",
			input: application.GetRestaurantDetailsInput{RestaurantID: "rest-1"},
			repo: &fakeRestaurantRepo{
				getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
					return entities.Restaurant{}, errors.New("db timeout")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewGetRestaurantDetailsUseCase(tc.repo)

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
			if got.ID != tc.wantID {
				t.Errorf("expected id %q, got %q", tc.wantID, got.ID)
			}
		})
	}
}
