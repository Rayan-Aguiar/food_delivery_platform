package httpdelivery

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apperrors "food_delivery_platform/shared/errors"
	"food_delivery_platform/shared/utils"

	"food_delivery_platform/services/restaurant-service/internal/application"
)

type createRestaurantRequest struct {
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	PhoneNumber string  `json:"phone_number"`
	Status      string  `json:"status,omitempty"`
	DeliveryFee float64 `json:"delivery_fee"`
}

type createMenuItemRequest struct {
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Category  string  `json:"category"`
	Available *bool   `json:"available,omitempty"`
}

type updateMenuItemRequest struct {
	Name      *string  `json:"name,omitempty"`
	Price     *float64 `json:"price,omitempty"`
	Category  *string  `json:"category,omitempty"`
	Available *bool    `json:"available,omitempty"`
}

func (h *RestaurantHandlers) CreateRestaurantHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.createRestaurant == nil {
		writeAppError(w, r, apperrors.Internal("create restaurant use case is not configured", nil))
		return
	}

	var req createRestaurantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAppError(w, r, apperrors.InvalidArgument("invalid request body", err))
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	restaurant, err := h.createRestaurant.Execute(r.Context(), application.CreateRestaurantInput{
		Name:           req.Name,
		Address:        req.Address,
		PhoneNumber:    req.PhoneNumber,
		Status:         status,
		DeliveryFee:    req.DeliveryFee,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, toRestaurantResponse(restaurant))
}

func (h *RestaurantHandlers) CreateMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.createMenuItem == nil {
		writeAppError(w, r, apperrors.Internal("create menu item use case is not configured", nil))
		return
	}

	var req createMenuItemRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAppError(w, r, apperrors.InvalidArgument("invalid request body", err))
		return
	}

	available := true
	if req.Available != nil {
		available = *req.Available
	}

	item, err := h.createMenuItem.Execute(r.Context(), application.CreateMenuItemInput{
		RestaurantID:   r.PathValue("id"),
		Name:           req.Name,
		Price:          req.Price,
		Category:       req.Category,
		Available:      available,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, toMenuItemResponse(item))
}

func (h *RestaurantHandlers) UpdateMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.updateMenuItem == nil {
		writeAppError(w, r, apperrors.Internal("update menu item use case is not configured", nil))
		return
	}

	var req updateMenuItemRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAppError(w, r, apperrors.InvalidArgument("invalid request body", err))
		return
	}

	item, err := h.updateMenuItem.Execute(r.Context(), application.UpdateMenuItemInput{
		RestaurantID: r.PathValue("id"),
		ItemID:       r.PathValue("itemId"),
		Name:         req.Name,
		Price:        req.Price,
		Category:     req.Category,
		Available:    req.Available,
	})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, toMenuItemResponse(item))
}

func decodeJSON(r *http.Request, dest any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}
