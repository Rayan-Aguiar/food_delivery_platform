package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

const (
	ValidationReasonItemNotFound       = "item_not_found"
	ValidationReasonItemUnavailable    = "item_unavailable"
	ValidationReasonRestaurantInactive = "restaurant_inactive"
)

type ValidateMenuItemsInput struct {
	RestaurantID string
	ItemIDs      []string
}

type MenuItemValidationResult struct {
	ItemID    string
	Found     bool
	Available bool
	Price     float64
	Valid     bool
	Reason    string
}

type ValidateMenuItemsOutput struct {
	RestaurantID     string
	RestaurantActive bool
	AllValid         bool
	Items            []MenuItemValidationResult
}

type ValidateMenuItemsUseCase struct {
	restaurants ports.RestaurantRepository
	menu        ports.MenuRepository
}

func NewValidateMenuItemsUseCase(restaurants ports.RestaurantRepository, menu ports.MenuRepository) *ValidateMenuItemsUseCase {
	return &ValidateMenuItemsUseCase{restaurants: restaurants, menu: menu}
}

func (uc *ValidateMenuItemsUseCase) Execute(ctx context.Context, input ValidateMenuItemsInput) (ValidateMenuItemsOutput, error) {
	if input.RestaurantID == "" {
		return ValidateMenuItemsOutput{}, apperrors.InvalidArgument("restaurant id is required", nil)
	}
	if len(input.ItemIDs) == 0 {
		return ValidateMenuItemsOutput{}, apperrors.InvalidArgument("at least one item id is required", nil)
	}

	restaurant, err := uc.restaurants.GetByID(ctx, input.RestaurantID)
	if err != nil {
		return ValidateMenuItemsOutput{}, mapNotFoundOrInternal(err, domainerrors.ErrRestaurantNotFound, "restaurant not found", "failed to get restaurant")
	}

	// Busca o cardapio inteiro de uma vez (indexado por restaurant_id) em vez de uma
	// consulta por item: evita N+1 contra o Mongo neste hot path de criacao de pedido.
	menuItems, err := uc.menu.ListByRestaurantID(ctx, input.RestaurantID)
	if err != nil {
		return ValidateMenuItemsOutput{}, apperrors.Internal("failed to list restaurant menu", err)
	}

	byID := make(map[string]entities.MenuItem, len(menuItems))
	for _, item := range menuItems {
		byID[item.ID] = item
	}

	restaurantActive := restaurant.IsActive()
	allValid := restaurantActive
	results := make([]MenuItemValidationResult, 0, len(input.ItemIDs))

	for _, itemID := range input.ItemIDs {
		item, found := byID[itemID]
		if !found {
			allValid = false
			results = append(results, MenuItemValidationResult{
				ItemID: itemID,
				Found:  false,
				Valid:  false,
				Reason: ValidationReasonItemNotFound,
			})
			continue
		}

		reason := ""
		valid := restaurantActive && item.Available
		switch {
		case !restaurantActive:
			reason = ValidationReasonRestaurantInactive
		case !item.Available:
			reason = ValidationReasonItemUnavailable
		}
		if !valid {
			allValid = false
		}

		results = append(results, MenuItemValidationResult{
			ItemID:    itemID,
			Found:     true,
			Available: item.Available,
			Price:     item.Price,
			Valid:     valid,
			Reason:    reason,
		})
	}

	return ValidateMenuItemsOutput{
		RestaurantID:     input.RestaurantID,
		RestaurantActive: restaurantActive,
		AllValid:         allValid,
		Items:            results,
	}, nil
}
