package mongo

import (
	"testing"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

func TestMapRestaurantDocToEntity_Success(t *testing.T) {
	doc := restaurantDoc{
		ID:          "rest-1",
		Name:        "Pizza Place",
		Address:     "Main St 123",
		PhoneNumber: "+5511999999999",
		Status:      string(entities.RestaurantStatusActive),
		DeliveryFee: 5.5,
	}

	restaurant, err := mapRestaurantDocToEntity(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restaurant.ID != doc.ID {
		t.Errorf("ID = %q, want %q", restaurant.ID, doc.ID)
	}
	if restaurant.Name != doc.Name {
		t.Errorf("Name = %q, want %q", restaurant.Name, doc.Name)
	}
	if restaurant.Status != entities.RestaurantStatusActive {
		t.Errorf("Status = %q, want %q", restaurant.Status, entities.RestaurantStatusActive)
	}
	if restaurant.DeliveryFee != doc.DeliveryFee {
		t.Errorf("DeliveryFee = %v, want %v", restaurant.DeliveryFee, doc.DeliveryFee)
	}
}

func TestMapRestaurantDocToEntity_InvalidStatus(t *testing.T) {
	doc := restaurantDoc{
		ID:          "rest-1",
		Name:        "Pizza Place",
		Address:     "Main St 123",
		PhoneNumber: "+5511999999999",
		Status:      "unknown_status",
		DeliveryFee: 5.5,
	}

	_, err := mapRestaurantDocToEntity(doc)
	if err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestMapRestaurantDocToEntity_InvalidDeliveryFee(t *testing.T) {
	doc := restaurantDoc{
		ID:          "rest-1",
		Name:        "Pizza Place",
		Address:     "Main St 123",
		PhoneNumber: "+5511999999999",
		Status:      string(entities.RestaurantStatusActive),
		DeliveryFee: -1,
	}

	_, err := mapRestaurantDocToEntity(doc)
	if err == nil {
		t.Error("expected error for negative delivery fee, got nil")
	}
}
