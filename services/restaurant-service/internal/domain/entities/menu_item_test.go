package entities

import (
	"errors"
	"testing"

	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
)

func TestNewMenuItem_Success(t *testing.T) {
	item, err := NewMenuItem("item_1", "rest_1", "Burger", 29.9, MenuCategoryHamburgueria, true)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if item.ID != "item_1" {
		t.Fatalf("unexpected id: %s", item.ID)
	}
	if item.RestaurantID != "rest_1" {
		t.Fatalf("unexpected restaurant id: %s", item.RestaurantID)
	}
	if item.Name != "Burger" {
		t.Fatalf("unexpected name: %s", item.Name)
	}
}

func TestNewMenuItem_InvalidID(t *testing.T) {
	_, err := NewMenuItem("", "rest_1", "Burger", 29.9, MenuCategoryHamburgueria, true)
	if !errors.Is(err, domainerrors.ErrInvalidMenuItemID) {
		t.Fatalf("expected ErrInvalidMenuItemID, got %v", err)
	}
}

func TestNewMenuItem_InvalidRestaurantID(t *testing.T) {
	_, err := NewMenuItem("item_1", "", "Burger", 29.9, MenuCategoryHamburgueria, true)
	if !errors.Is(err, domainerrors.ErrMenuItemRestaurant) {
		t.Fatalf("expected ErrMenuItemRestaurant, got %v", err)
	}
}

func TestNewMenuItem_InvalidName(t *testing.T) {
	_, err := NewMenuItem("item_1", "rest_1", "", 29.9, MenuCategoryHamburgueria, true)
	if !errors.Is(err, domainerrors.ErrInvalidMenuItemName) {
		t.Fatalf("expected ErrInvalidMenuItemName, got %v", err)
	}
}

func TestNewMenuItem_InvalidCategory(t *testing.T) {
	_, err := NewMenuItem("item_1", "rest_1", "Burger", 29.9, "", true)
	if !errors.Is(err, domainerrors.ErrInvalidCategory) {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestNewMenuItem_UnsupportedCategory(t *testing.T) {
	_, err := NewMenuItem("item_1", "rest_1", "Burger", 29.9, MenuCategory("mexicana"), true)
	if !errors.Is(err, domainerrors.ErrInvalidCategory) {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestNewMenuItem_InvalidPrice(t *testing.T) {
	_, err := NewMenuItem("item_1", "rest_1", "Burger", 0, MenuCategoryHamburgueria, true)
	if !errors.Is(err, domainerrors.ErrInvalidPrice) {
		t.Fatalf("expected ErrInvalidPrice, got %v", err)
	}
}

func TestMenuItem_EnsureAvailable_WhenUnavailable(t *testing.T) {
	item, err := NewMenuItem("item_1", "rest_1", "Burger", 29.9, MenuCategoryHamburgueria, false)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	err = item.EnsureAvailable()
	if !errors.Is(err, domainerrors.ErrItemUnavailable) {
		t.Fatalf("expected ErrItemUnavailable, got %v", err)
	}
}

func TestMenuCategory_IsValid(t *testing.T) {
	validCategories := []MenuCategory{
		MenuCategoryPizzaria,
		MenuCategoryHamburgueria,
		MenuCategoryJapones,
		MenuCategoryComidaBrasileira,
		MenuCategorySorveteria,
	}

	for _, category := range validCategories {
		if !category.IsValid() {
			t.Fatalf("expected category %q to be valid", category)
		}
	}

	if MenuCategory("mexicana").IsValid() {
		t.Fatal("expected unsupported category to be invalid")
	}
}
