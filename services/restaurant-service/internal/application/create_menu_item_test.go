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

func TestCreateMenuItemUseCase_Execute_IdempotentReplay(t *testing.T) {
	ctx := context.Background()
	restaurants := &fakeRestaurantRepo{
		getByIDFn: func(_ context.Context, _ string) (entities.Restaurant, error) {
			return mustRestaurant(), nil
		},
	}
	input := application.CreateMenuItemInput{
		RestaurantID:   "rest-1",
		Name:           "Margherita (retry com body diferente)",
		Price:          99,
		Category:       string(entities.MenuCategoryPizzaria),
		Available:      true,
		IdempotencyKey: "client-key-1",
	}
	original, err := entities.NewMenuItem("client-key-1", "rest-1", "Margherita", 39.9, entities.MenuCategoryPizzaria, true)
	if err != nil {
		t.Fatalf("build original: %v", err)
	}

	saveDuplicate := func(context.Context, entities.MenuItem) error { return domainerrors.ErrMenuItemExists }

	t.Run("retry returns the originally created item without republishing", func(t *testing.T) {
		publisher := &fakeCatalogEventPublisher{}
		menu := &fakeMenuRepo{
			saveFn: saveDuplicate,
			getByIDFn: func(_ context.Context, restaurantID, itemID string) (entities.MenuItem, error) {
				if restaurantID != "rest-1" || itemID != "client-key-1" {
					t.Errorf("lookup = (%q, %q), want (rest-1, client-key-1)", restaurantID, itemID)
				}
				return original, nil
			},
		}

		got, err := application.NewCreateMenuItemUseCase(restaurants, menu, &fakeIDGenerator{}, publisher).Execute(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != original {
			t.Errorf("replay = %+v, want original %+v (retry nao sobrescreve)", got, original)
		}
		if len(publisher.menuUpdatedCalls) != 0 {
			t.Errorf("event published %d times on replay, want 0", len(publisher.menuUpdatedCalls))
		}
	})

	t.Run("key used by another restaurant returns conflict", func(t *testing.T) {
		menu := &fakeMenuRepo{
			saveFn: saveDuplicate,
			getByIDFn: func(context.Context, string, string) (entities.MenuItem, error) {
				return entities.MenuItem{}, domainerrors.ErrMenuItemNotFound
			},
		}

		_, err := application.NewCreateMenuItemUseCase(restaurants, menu, &fakeIDGenerator{}).Execute(ctx, input)

		var appErr *apperrors.AppError
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeConflict {
			t.Fatalf("expected conflict, got %v", err)
		}
	})

	t.Run("duplicate without idempotency key is an internal error", func(t *testing.T) {
		noKey := input
		noKey.IdempotencyKey = ""
		menu := &fakeMenuRepo{saveFn: saveDuplicate}

		_, err := application.NewCreateMenuItemUseCase(restaurants, menu, &fakeIDGenerator{ids: []string{"gen-1"}}).Execute(ctx, noKey)

		var appErr *apperrors.AppError
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInternal {
			t.Fatalf("expected internal error, got %v", err)
		}
	})
}
