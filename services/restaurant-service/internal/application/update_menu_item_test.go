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

func strPtr(s string) *string     { return &s }
func floatPtr(f float64) *float64 { return &f }
func boolPtr(b bool) *bool        { return &b }

func TestUpdateMenuItemUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	itemFound := &fakeMenuRepo{
		getByIDFn: func(_ context.Context, restaurantID, itemID string) (entities.MenuItem, error) {
			return mustMenuItem(), nil
		},
	}

	tests := []struct {
		name          string
		input         application.UpdateMenuItemInput
		menu          *fakeMenuRepo
		wantAvailable bool
		wantPrice     float64
		wantErr       bool
		wantCode      string
	}{
		{
			name: "toggles availability only",
			input: application.UpdateMenuItemInput{
				RestaurantID: "rest-1",
				ItemID:       "item-1",
				Available:    boolPtr(false),
			},
			menu:          itemFound,
			wantAvailable: false,
			wantPrice:     39.9,
		},
		{
			name: "updates price only",
			input: application.UpdateMenuItemInput{
				RestaurantID: "rest-1",
				ItemID:       "item-1",
				Price:        floatPtr(49.9),
			},
			menu:          itemFound,
			wantAvailable: true,
			wantPrice:     49.9,
		},
		{
			name:     "missing ids",
			input:    application.UpdateMenuItemInput{},
			menu:     &fakeMenuRepo{},
			wantErr:  true,
			wantCode: apperrors.CodeInvalidArgument,
		},
		{
			name:  "item not found",
			input: application.UpdateMenuItemInput{RestaurantID: "rest-1", ItemID: "missing"},
			menu: &fakeMenuRepo{
				getByIDFn: func(_ context.Context, restaurantID, itemID string) (entities.MenuItem, error) {
					return entities.MenuItem{}, domainerrors.ErrMenuItemNotFound
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeNotFound,
		},
		{
			name: "invalid price maps to invalid argument",
			input: application.UpdateMenuItemInput{
				RestaurantID: "rest-1",
				ItemID:       "item-1",
				Price:        floatPtr(-1),
			},
			menu:     itemFound,
			wantErr:  true,
			wantCode: apperrors.CodeInvalidArgument,
		},
		{
			name: "repository update error maps to internal",
			input: application.UpdateMenuItemInput{
				RestaurantID: "rest-1",
				ItemID:       "item-1",
				Name:         strPtr("New Name"),
			},
			menu: &fakeMenuRepo{
				getByIDFn: itemFound.getByIDFn,
				updateFn: func(_ context.Context, _ entities.MenuItem) error {
					return errors.New("db unavailable")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewUpdateMenuItemUseCase(tc.menu)

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
			if got.Available != tc.wantAvailable {
				t.Errorf("available = %v, want %v", got.Available, tc.wantAvailable)
			}
			if got.Price != tc.wantPrice {
				t.Errorf("price = %v, want %v", got.Price, tc.wantPrice)
			}
		})
	}
}
