package application_test

import (
	"context"
	"errors"
	"testing"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

func TestCreateRestaurantUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	validInput := application.CreateRestaurantInput{
		Name:        "Pizza Place",
		Address:     "Main St 123",
		PhoneNumber: "+5511999999999",
		Status:      string(entities.RestaurantStatusActive),
		DeliveryFee: 5.0,
	}

	tests := []struct {
		name     string
		input    application.CreateRestaurantInput
		repo     *fakeRestaurantRepo
		idGen    *fakeIDGenerator
		wantID   string
		wantErr  bool
		wantCode string
	}{
		{
			name:   "success generates id",
			input:  validInput,
			repo:   &fakeRestaurantRepo{},
			idGen:  &fakeIDGenerator{ids: []string{"generated-1"}},
			wantID: "generated-1",
		},
		{
			name: "idempotency key is used as id",
			input: application.CreateRestaurantInput{
				Name:           validInput.Name,
				Address:        validInput.Address,
				PhoneNumber:    validInput.PhoneNumber,
				Status:         validInput.Status,
				DeliveryFee:    validInput.DeliveryFee,
				IdempotencyKey: "client-key-1",
			},
			repo:   &fakeRestaurantRepo{},
			idGen:  &fakeIDGenerator{},
			wantID: "client-key-1",
		},
		{
			name:     "invalid name maps to invalid argument",
			input:    application.CreateRestaurantInput{Address: "x", PhoneNumber: "y", Status: string(entities.RestaurantStatusActive)},
			repo:     &fakeRestaurantRepo{},
			idGen:    &fakeIDGenerator{},
			wantErr:  true,
			wantCode: apperrors.CodeInvalidArgument,
		},
		{
			name:  "repository save error maps to internal",
			input: validInput,
			repo: &fakeRestaurantRepo{
				saveFn: func(_ context.Context, _ entities.Restaurant) error {
					return errors.New("db unavailable")
				},
			},
			idGen:    &fakeIDGenerator{ids: []string{"generated-1"}},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewCreateRestaurantUseCase(tc.repo, tc.idGen)

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
