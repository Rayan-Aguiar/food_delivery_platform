package httpdelivery

import (
	"context"
	"net/http"

	apperrors "food_delivery_platform/shared/errors"
	"food_delivery_platform/shared/utils"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
)

type listRestaurantsExecutor interface {
	Execute(ctx context.Context) ([]entities.Restaurant, error)
}

type getRestaurantDetailsExecutor interface {
	Execute(ctx context.Context, input application.GetRestaurantDetailsInput) (entities.Restaurant, error)
}

type listRestaurantMenuExecutor interface {
	Execute(ctx context.Context, input application.ListRestaurantMenuInput) ([]entities.MenuItem, error)
}

type createRestaurantExecutor interface {
	Execute(ctx context.Context, input application.CreateRestaurantInput) (entities.Restaurant, error)
}

type createMenuItemExecutor interface {
	Execute(ctx context.Context, input application.CreateMenuItemInput) (entities.MenuItem, error)
}

type updateMenuItemExecutor interface {
	Execute(ctx context.Context, input application.UpdateMenuItemInput) (entities.MenuItem, error)
}

type RestaurantHandlers struct {
	listRestaurants  listRestaurantsExecutor
	getDetails       getRestaurantDetailsExecutor
	listMenu         listRestaurantMenuExecutor
	createRestaurant createRestaurantExecutor
	createMenuItem   createMenuItemExecutor
	updateMenuItem   updateMenuItemExecutor
}

func NewRestaurantHandlers(
	listRestaurants listRestaurantsExecutor,
	getDetails getRestaurantDetailsExecutor,
	listMenu listRestaurantMenuExecutor,
	createRestaurant createRestaurantExecutor,
	createMenuItem createMenuItemExecutor,
	updateMenuItem updateMenuItemExecutor,
) *RestaurantHandlers {
	return &RestaurantHandlers{
		listRestaurants:  listRestaurants,
		getDetails:       getDetails,
		listMenu:         listMenu,
		createRestaurant: createRestaurant,
		createMenuItem:   createMenuItem,
		updateMenuItem:   updateMenuItem,
	}
}

type restaurantResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	PhoneNumber string  `json:"phone_number"`
	Status      string  `json:"status"`
	DeliveryFee float64 `json:"delivery_fee"`
}

type listRestaurantsResponse struct {
	Restaurants []restaurantResponse `json:"restaurants"`
}

type menuItemResponse struct {
	ID           string  `json:"id"`
	RestaurantID string  `json:"restaurant_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Category     string  `json:"category"`
	Available    bool    `json:"available"`
}

type listRestaurantMenuResponse struct {
	Items []menuItemResponse `json:"items"`
}

func (h *RestaurantHandlers) ListRestaurantsHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.listRestaurants == nil {
		writeAppError(w, r, apperrors.Internal("list restaurants use case is not configured", nil))
		return
	}

	restaurants, err := h.listRestaurants.Execute(r.Context())
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, listRestaurantsResponse{
		Restaurants: toRestaurantResponses(restaurants),
	})
}

func (h *RestaurantHandlers) GetRestaurantDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.getDetails == nil {
		writeAppError(w, r, apperrors.Internal("get restaurant details use case is not configured", nil))
		return
	}

	restaurant, err := h.getDetails.Execute(r.Context(), application.GetRestaurantDetailsInput{
		RestaurantID: r.PathValue("id"),
	})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, toRestaurantResponse(restaurant))
}

func (h *RestaurantHandlers) ListRestaurantMenuHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.listMenu == nil {
		writeAppError(w, r, apperrors.Internal("list restaurant menu use case is not configured", nil))
		return
	}

	items, err := h.listMenu.Execute(r.Context(), application.ListRestaurantMenuInput{
		RestaurantID: r.PathValue("id"),
	})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, listRestaurantMenuResponse{
		Items: toMenuItemResponses(items),
	})
}

func toRestaurantResponse(r entities.Restaurant) restaurantResponse {
	return restaurantResponse{
		ID:          r.ID,
		Name:        r.Name,
		Address:     r.Address,
		PhoneNumber: r.PhoneNumber,
		Status:      string(r.Status),
		DeliveryFee: r.DeliveryFee,
	}
}

func toRestaurantResponses(restaurants []entities.Restaurant) []restaurantResponse {
	out := make([]restaurantResponse, 0, len(restaurants))
	for _, r := range restaurants {
		out = append(out, toRestaurantResponse(r))
	}
	return out
}

func toMenuItemResponse(item entities.MenuItem) menuItemResponse {
	return menuItemResponse{
		ID:           item.ID,
		RestaurantID: item.RestaurantID,
		Name:         item.Name,
		Price:        item.Price,
		Category:     string(item.Category),
		Available:    item.Available,
	}
}

func toMenuItemResponses(items []entities.MenuItem) []menuItemResponse {
	out := make([]menuItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toMenuItemResponse(item))
	}
	return out
}
