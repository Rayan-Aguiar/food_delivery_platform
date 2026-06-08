package entities

import (
	"errors"
	"testing"

	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
)

func TestNewRestaurant_Success(t *testing.T) {
	restaurant, err := NewRestaurant(
		"rest_1",
		"Pizza Prime",
		"Av. Paulista, 1000",
		"11999999999",
		RestaurantStatusActive,
		8.5,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if restaurant.ID != "rest_1" {
		t.Fatalf("unexpected id: %s", restaurant.ID)
	}
	if restaurant.Name != "Pizza Prime" {
		t.Fatalf("unexpected name: %s", restaurant.Name)
	}
	if restaurant.Address != "Av. Paulista, 1000" {
		t.Fatalf("unexpected address: %s", restaurant.Address)
	}
	if restaurant.PhoneNumber != "11999999999" {
		t.Fatalf("unexpected phone number: %s", restaurant.PhoneNumber)
	}
	if !restaurant.IsActive() {
		t.Fatal("expected restaurant to be active")
	}
}

func TestNewRestaurant_InvalidID(t *testing.T) {
	_, err := NewRestaurant("", "Pizza Prime", "Av. Paulista, 1000", "11999999999", RestaurantStatusActive, 8.5)
	if !errors.Is(err, domainerrors.ErrInvalidRestaurantID) {
		t.Fatalf("expected ErrInvalidRestaurantID, got %v", err)
	}
}

func TestNewRestaurant_InvalidName(t *testing.T) {
	_, err := NewRestaurant("rest_1", "", "Av. Paulista, 1000", "11999999999", RestaurantStatusActive, 8.5)
	if !errors.Is(err, domainerrors.ErrInvalidRestaurantName) {
		t.Fatalf("expected ErrInvalidRestaurantName, got %v", err)
	}
}

func TestNewRestaurant_InvalidAddress(t *testing.T) {
	_, err := NewRestaurant("rest_1", "Pizza Prime", "", "11999999999", RestaurantStatusActive, 8.5)
	if !errors.Is(err, domainerrors.ErrInvalidRestaurantAddress) {
		t.Fatalf("expected ErrInvalidRestaurantAddress, got %v", err)
	}
}

func TestNewRestaurant_InvalidPhone(t *testing.T) {
	_, err := NewRestaurant("rest_1", "Pizza Prime", "Av. Paulista, 1000", "", RestaurantStatusActive, 8.5)
	if !errors.Is(err, domainerrors.ErrInvalidRestaurantPhone) {
		t.Fatalf("expected ErrInvalidRestaurantPhone, got %v", err)
	}
}

func TestNewRestaurant_InvalidStatus(t *testing.T) {
	_, err := NewRestaurant("rest_1", "Pizza Prime", "Av. Paulista, 1000", "11999999999", RestaurantStatus("unknown"), 8.5)
	if !errors.Is(err, domainerrors.ErrInvalidRestaurantStatus) {
		t.Fatalf("expected ErrInvalidRestaurantStatus, got %v", err)
	}
}

func TestNewRestaurant_InvalidDeliveryFee(t *testing.T) {
	_, err := NewRestaurant("rest_1", "Pizza Prime", "Av. Paulista, 1000", "11999999999", RestaurantStatusActive, -1)
	if !errors.Is(err, domainerrors.ErrInvalidDeliveryFee) {
		t.Fatalf("expected ErrInvalidDeliveryFee, got %v", err)
	}
}

func TestRestaurant_EnsureAcceptingOrders_WhenInactive(t *testing.T) {
	restaurant, err := NewRestaurant(
		"rest_1",
		"Pizza Prime",
		"Av. Paulista, 1000",
		"11999999999",
		RestaurantStatusInactive,
		8.5,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	err = restaurant.EnsureAcceptingOrders()
	if !errors.Is(err, domainerrors.ErrRestaurantInactive) {
		t.Fatalf("expected ErrRestaurantInactive, got %v", err)
	}
}
