package entities

import (
	"strings"

	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
)

type RestaurantStatus string

const (
	RestaurantStatusActive   RestaurantStatus = "active"
	RestaurantStatusInactive RestaurantStatus = "inactive"
)

type Restaurant struct {
	ID          string
	Name        string
	Address     string
	PhoneNumber string
	Status      RestaurantStatus
	DeliveryFee float64
}

func NewRestaurant(id, name, address, phoneNumber string, status RestaurantStatus, deliveryFee float64) (Restaurant, error) {
	restaurant := Restaurant{
		ID:          strings.TrimSpace(id),
		Name:        strings.TrimSpace(name),
		Address:     strings.TrimSpace(address),
		PhoneNumber: strings.TrimSpace(phoneNumber),
		Status:      status,
		DeliveryFee: deliveryFee,
	}

	if err := restaurant.Validate(); err != nil {
		return Restaurant{}, err
	}

	return restaurant, nil
}

func (r Restaurant) Validate() error {
	if r.ID == "" {
		return domainerrors.ErrInvalidRestaurantID
	}
	if r.Name == "" {
		return domainerrors.ErrInvalidRestaurantName
	}
	if r.Address == "" {
		return domainerrors.ErrInvalidRestaurantAddress
	}
	if r.PhoneNumber == "" {
		return domainerrors.ErrInvalidRestaurantPhone
	}
	if !r.Status.IsValid() {
		return domainerrors.ErrInvalidRestaurantStatus
	}
	if r.DeliveryFee < 0 {
		return domainerrors.ErrInvalidDeliveryFee
	}

	return nil
}

func (r Restaurant) IsActive() bool {
	return r.Status == RestaurantStatusActive
}

func (r Restaurant) EnsureAcceptingOrders() error {
	if !r.IsActive() {
		return domainerrors.ErrRestaurantInactive
	}
	return nil
}

func (s RestaurantStatus) IsValid() bool {
	return s == RestaurantStatusActive || s == RestaurantStatusInactive
}
