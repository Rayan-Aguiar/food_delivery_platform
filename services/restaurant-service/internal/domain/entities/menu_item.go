package entities

import (
	"strings"

	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
)

type MenuCategory string

const (
	MenuCategoryPizzaria         MenuCategory = "pizzaria"
	MenuCategoryHamburgueria     MenuCategory = "hamburgueria"
	MenuCategoryJapones          MenuCategory = "japones"
	MenuCategoryComidaBrasileira MenuCategory = "comida brasileira"
	MenuCategorySorveteria       MenuCategory = "sorveteria"
)

type MenuItem struct {
	ID           string
	RestaurantID string
	Name         string
	Price        float64
	Category     MenuCategory
	Available    bool
}

func NewMenuItem(id, restaurantID, name string, price float64, category MenuCategory, available bool) (MenuItem, error) {
	item := MenuItem{
		ID:           strings.TrimSpace(id),
		RestaurantID: strings.TrimSpace(restaurantID),
		Name:         strings.TrimSpace(name),
		Price:        price,
		Category:     MenuCategory(strings.TrimSpace(string(category))),
		Available:    available,
	}

	if err := item.Validate(); err != nil {
		return MenuItem{}, err
	}

	return item, nil
}

func (m MenuItem) Validate() error {
	if m.ID == "" {
		return domainerrors.ErrInvalidMenuItemID
	}

	if m.RestaurantID == "" {
		return domainerrors.ErrMenuItemRestaurant
	}

	if m.Name == "" {
		return domainerrors.ErrInvalidMenuItemName
	}

	if m.Category == "" {
		return domainerrors.ErrInvalidCategory
	}

	if !m.Category.IsValid() {
		return domainerrors.ErrInvalidCategory
	}

	if m.Price <= 0 {
		return domainerrors.ErrInvalidPrice
	}

	return nil
}

func (m MenuItem) EnsureAvailable() error {
	if !m.Available {
		return domainerrors.ErrItemUnavailable
	}
	return nil
}

func (c MenuCategory) IsValid() bool {
	return c == MenuCategoryPizzaria ||
		c == MenuCategoryHamburgueria ||
		c == MenuCategoryJapones ||
		c == MenuCategoryComidaBrasileira ||
		c == MenuCategorySorveteria
}
