package domainerrors

import "errors"

var (
	ErrInvalidRestaurantID     = errors.New("invalid restaurant id")
	ErrInvalidRestaurantName   = errors.New("invalid restaurant name")
	ErrInvalidRestaurantAddress = errors.New("invalid restaurant address")
	ErrInvalidRestaurantPhone  = errors.New("invalid restaurant phone")
	ErrInvalidDeliveryFee      = errors.New("invalid delivery fee")
	ErrInvalidRestaurantStatus = errors.New("invalid restaurant status")
	ErrRestaurantInactive      = errors.New("restaurant is inactive")

	ErrInvalidMenuItemID   = errors.New("invalid menu item id")
	ErrInvalidMenuItemName = errors.New("invalid menu item name")
	ErrInvalidCategory     = errors.New("invalid category")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrItemUnavailable     = errors.New("menu item is unavailable")
	ErrMenuItemRestaurant  = errors.New("menu item restaurant id is invalid")
)
