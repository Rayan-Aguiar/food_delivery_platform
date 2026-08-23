package application_test

import (
	"context"
	"errors"
	"testing"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

func TestCreateMenuItemUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	restaurantFound := &fakeRestaurantRepo{
		getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
			return mustRestaurant(), nil
		},
	}
	validInput := application.CreateMenuItemInput{
		RestaurantID: "rest-1",
		Name:         "Margherita",
		Price:        39.9,
		Category:     string(entities.MenuCategoryPizzaria),
		Available:    true,
	}

	tests := []struct {
		name        string
		input       application.CreateMenuItemInput
		restaurants *fakeRestaurantRepo
		menu        *fakeMenuRepo
		idGen       *fakeIDGenerator
		wantID      string
		wantErr     bool
		wantCode    string
	}{
		{
			name:        "success generates id",
			input:       validInput,
			restaurants: restaurantFound,
			menu:        &fakeMenuRepo{},
			idGen:       &fakeIDGenerator{ids: []string{"generated-item-1"}},
			wantID:      "generated-item-1",
		},
		{
			name: "idempotency key is used as id",
			input: application.CreateMenuItemInput{
				RestaurantID:   validInput.RestaurantID,
				Name:           validInput.Name,
				Price:          validInput.Price,
				Category:       validInput.Category,
				Available:      validInput.Available,
				IdempotencyKey: "client-key-1",
			},
			restaurants: restaurantFound,
			menu:        &fakeMenuRepo{},
			idGen:       &fakeIDGenerator{},
			wantID:      "client-key-1",
		},
		{
			name:        "empty restaurant id",
			input:       application.CreateMenuItemInput{RestaurantID: ""},
			restaurants: &fakeRestaurantRepo{},
			menu:        &fakeMenuRepo{},
			idGen:       &fakeIDGenerator{},
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
			idGen:    &fakeIDGenerator{},
			wantErr:  true,
			wantCode: apperrors.CodeNotFound,
		},
		{
			name:        "invalid category maps to invalid argument",
			input:       application.CreateMenuItemInput{RestaurantID: "rest-1", Name: "X", Price: 10, Category: "unknown"},
			restaurants: restaurantFound,
			menu:        &fakeMenuRepo{},
			idGen:       &fakeIDGenerator{},
			wantErr:     true,
			wantCode:    apperrors.CodeInvalidArgument,
		},
		{
			name:        "repository save error maps to internal",
			input:       validInput,
			restaurants: restaurantFound,
			menu: &fakeMenuRepo{
				saveFn: func(_ context.Context, _ entities.MenuItem) error {
					return errors.New("db unavailable")
				},
			},
			idGen:    &fakeIDGenerator{ids: []string{"generated-item-1"}},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewCreateMenuItemUseCase(tc.restaurants, tc.menu, tc.idGen)

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

func TestCreateMenuItemUseCase_PublishesMenuUpdatedEvent(t *testing.T) {
	ctx := context.Background()
	restaurantFound := &fakeRestaurantRepo{
		getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
			return mustRestaurant(), nil
		},
	}
	eventsPublisher := &fakeCatalogEventPublisher{}

	uc := application.NewCreateMenuItemUseCase(restaurantFound, &fakeMenuRepo{}, &fakeIDGenerator{ids: []string{"item-x"}}, eventsPublisher)

	_, err := uc.Execute(ctx, application.CreateMenuItemInput{
		RestaurantID:  "rest-1",
		Name:          "Margherita",
		Price:         39.9,
		Category:      string(entities.MenuCategoryPizzaria),
		Available:     true,
		CorrelationID: "corr-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(eventsPublisher.menuUpdatedCalls) != 1 {
		t.Fatalf("expected 1 published menu updated event, got %d", len(eventsPublisher.menuUpdatedCalls))
	}
	got := eventsPublisher.menuUpdatedCalls[0]
	if got.ItemID != "item-x" || got.RestaurantID != "rest-1" {
		t.Errorf("unexpected event: %+v", got)
	}
	if got.CorrelationID != "corr-1" {
		t.Errorf("correlation id = %q, want %q", got.CorrelationID, "corr-1")
	}
}

func TestCreateMenuItemUseCase_PublishFailureIsBestEffort(t *testing.T) {
	ctx := context.Background()
	restaurantFound := &fakeRestaurantRepo{
		getByIDFn: func(_ context.Context, id string) (entities.Restaurant, error) {
			return mustRestaurant(), nil
		},
	}
	eventsPublisher := &fakeCatalogEventPublisher{
		publishMenuUpdatedFn: func(context.Context, ports.MenuUpdatedEvent) error {
			return errors.New("broker unavailable")
		},
	}

	uc := application.NewCreateMenuItemUseCase(restaurantFound, &fakeMenuRepo{}, &fakeIDGenerator{}, eventsPublisher)

	got, err := uc.Execute(ctx, application.CreateMenuItemInput{
		RestaurantID: "rest-1",
		Name:         "Margherita",
		Price:        39.9,
		Category:     string(entities.MenuCategoryPizzaria),
		Available:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID == "" {
		t.Fatal("expected successful creation output")
	}
	if len(eventsPublisher.menuUpdatedCalls) != 1 {
		t.Fatalf("expected 1 publish attempt, got %d", len(eventsPublisher.menuUpdatedCalls))
	}
}
