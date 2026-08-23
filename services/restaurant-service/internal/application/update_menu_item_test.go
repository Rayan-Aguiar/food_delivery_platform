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

func TestUpdateMenuItemUseCase_PublishesAvailabilityChangedWhenAvailabilityFlips(t *testing.T) {
	ctx := context.Background()
	menu := &fakeMenuRepo{
		getByIDFn: func(_ context.Context, _, _ string) (entities.MenuItem, error) {
			return mustMenuItem(), nil
		},
	}
	eventsPublisher := &fakeCatalogEventPublisher{}

	uc := application.NewUpdateMenuItemUseCase(menu, eventsPublisher)
	_, err := uc.Execute(ctx, application.UpdateMenuItemInput{
		RestaurantID:  "rest-1",
		ItemID:        "item-1",
		Available:     boolPtr(false),
		CorrelationID: "corr-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(eventsPublisher.availabilityChangedCalls) != 1 {
		t.Fatalf("expected 1 availability changed event, got %d", len(eventsPublisher.availabilityChangedCalls))
	}
	got := eventsPublisher.availabilityChangedCalls[0]
	if got.Available {
		t.Error("expected available = false in event")
	}
	if got.CorrelationID != "corr-1" {
		t.Errorf("correlation id = %q, want %q", got.CorrelationID, "corr-1")
	}
	if len(eventsPublisher.menuUpdatedCalls) != 0 {
		t.Errorf("expected no menu updated event, got %d", len(eventsPublisher.menuUpdatedCalls))
	}
}

func TestUpdateMenuItemUseCase_PublishesMenuUpdatedWhenPriceChanges(t *testing.T) {
	ctx := context.Background()
	menu := &fakeMenuRepo{
		getByIDFn: func(_ context.Context, _, _ string) (entities.MenuItem, error) {
			return mustMenuItem(), nil
		},
	}
	eventsPublisher := &fakeCatalogEventPublisher{}

	uc := application.NewUpdateMenuItemUseCase(menu, eventsPublisher)
	_, err := uc.Execute(ctx, application.UpdateMenuItemInput{
		RestaurantID: "rest-1",
		ItemID:       "item-1",
		Price:        floatPtr(49.9),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(eventsPublisher.menuUpdatedCalls) != 1 {
		t.Fatalf("expected 1 menu updated event, got %d", len(eventsPublisher.menuUpdatedCalls))
	}
	if eventsPublisher.menuUpdatedCalls[0].Price != 49.9 {
		t.Errorf("price in event = %v, want 49.9", eventsPublisher.menuUpdatedCalls[0].Price)
	}
	if len(eventsPublisher.availabilityChangedCalls) != 0 {
		t.Errorf("expected no availability changed event, got %d", len(eventsPublisher.availabilityChangedCalls))
	}
}

func TestUpdateMenuItemUseCase_NoEventWhenNothingChanges(t *testing.T) {
	ctx := context.Background()
	menu := &fakeMenuRepo{
		getByIDFn: func(_ context.Context, _, _ string) (entities.MenuItem, error) {
			return mustMenuItem(), nil
		},
	}
	eventsPublisher := &fakeCatalogEventPublisher{}

	uc := application.NewUpdateMenuItemUseCase(menu, eventsPublisher)
	item := mustMenuItem()
	_, err := uc.Execute(ctx, application.UpdateMenuItemInput{
		RestaurantID: "rest-1",
		ItemID:       "item-1",
		Available:    boolPtr(item.Available),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(eventsPublisher.menuUpdatedCalls) != 0 || len(eventsPublisher.availabilityChangedCalls) != 0 {
		t.Errorf("expected no events when nothing changed, got menuUpdated=%d availabilityChanged=%d",
			len(eventsPublisher.menuUpdatedCalls), len(eventsPublisher.availabilityChangedCalls))
	}
}

func TestUpdateMenuItemUseCase_PublishFailureIsBestEffort(t *testing.T) {
	ctx := context.Background()
	menu := &fakeMenuRepo{
		getByIDFn: func(_ context.Context, _, _ string) (entities.MenuItem, error) {
			return mustMenuItem(), nil
		},
	}
	eventsPublisher := &fakeCatalogEventPublisher{
		publishAvailabilityChangedFn: func(context.Context, ports.AvailabilityChangedEvent) error {
			return errors.New("broker unavailable")
		},
	}

	uc := application.NewUpdateMenuItemUseCase(menu, eventsPublisher)
	got, err := uc.Execute(ctx, application.UpdateMenuItemInput{
		RestaurantID: "rest-1",
		ItemID:       "item-1",
		Available:    boolPtr(false),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Available {
		t.Fatal("expected update to still succeed with available = false")
	}
	if len(eventsPublisher.availabilityChangedCalls) != 1 {
		t.Fatalf("expected 1 publish attempt, got %d", len(eventsPublisher.availabilityChangedCalls))
	}
}
