package mongo

import (
	"testing"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

func TestMapMenuItemDocToEntity_Success(t *testing.T) {
	doc := menuItemDoc{
		ID:           "item-1",
		RestaurantID: "rest-1",
		Name:         "Margherita",
		Price:        39.9,
		Category:     string(entities.MenuCategoryPizzaria),
		Available:    true,
	}

	item, err := mapMenuItemDocToEntity(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.ID != doc.ID {
		t.Errorf("ID = %q, want %q", item.ID, doc.ID)
	}
	if item.RestaurantID != doc.RestaurantID {
		t.Errorf("RestaurantID = %q, want %q", item.RestaurantID, doc.RestaurantID)
	}
	if item.Category != entities.MenuCategoryPizzaria {
		t.Errorf("Category = %q, want %q", item.Category, entities.MenuCategoryPizzaria)
	}
	if item.Available != doc.Available {
		t.Errorf("Available = %v, want %v", item.Available, doc.Available)
	}
}

func TestMapMenuItemDocToEntity_InvalidCategory(t *testing.T) {
	doc := menuItemDoc{
		ID:           "item-1",
		RestaurantID: "rest-1",
		Name:         "Margherita",
		Price:        39.9,
		Category:     "unknown_category",
		Available:    true,
	}

	_, err := mapMenuItemDocToEntity(doc)
	if err == nil {
		t.Error("expected error for invalid category, got nil")
	}
}

func TestMapMenuItemEntityToDoc(t *testing.T) {
	item, err := entities.NewMenuItem("item-1", "rest-1", "Margherita", 39.9, entities.MenuCategoryPizzaria, true)
	if err != nil {
		t.Fatalf("unexpected error building entity: %v", err)
	}

	doc := mapMenuItemEntityToDoc(item)

	if doc.ID != item.ID {
		t.Errorf("ID = %q, want %q", doc.ID, item.ID)
	}
	if doc.Category != string(item.Category) {
		t.Errorf("Category = %q, want %q", doc.Category, string(item.Category))
	}
}
