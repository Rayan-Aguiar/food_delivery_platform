package httpdelivery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	apperrors "food_delivery_platform/shared/errors"
)

func newAdminTestHandlers(create *fakeCreateRestaurant, createItem *fakeCreateMenuItem, updateItem *fakeUpdateMenuItem) *RestaurantHandlers {
	if create == nil {
		create = &fakeCreateRestaurant{}
	}
	if createItem == nil {
		createItem = &fakeCreateMenuItem{}
	}
	if updateItem == nil {
		updateItem = &fakeUpdateMenuItem{}
	}
	return NewRestaurantHandlers(&fakeListRestaurants{}, &fakeGetRestaurantDetails{}, &fakeListRestaurantMenu{}, create, createItem, updateItem, &fakeValidateMenuItems{})
}

// ── CreateRestaurantHandler ─────────────────────────────────────────────────────

func TestCreateRestaurantHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		create     *fakeCreateRestaurant
		wantStatus int
		wantCode   string
	}{
		{
			name:       "success returns 201",
			body:       `{"name":"Pizza Place","address":"Main St 123","phone_number":"+5511999999999","delivery_fee":5.0}`,
			create:     &fakeCreateRestaurant{},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json returns 400",
			body:       `{bad json`,
			create:     &fakeCreateRestaurant{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name:       "unknown field returns 400",
			body:       `{"name":"x","address":"y","phone_number":"z","delivery_fee":1,"extra":"nope"}`,
			create:     &fakeCreateRestaurant{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name: "domain validation error returns 400",
			body: `{"name":"","address":"y","phone_number":"z","delivery_fee":1}`,
			create: &fakeCreateRestaurant{
				fn: func(_ context.Context, _ application.CreateRestaurantInput) (entities.Restaurant, error) {
					return entities.Restaurant{}, apperrors.InvalidArgument("invalid restaurant name", nil)
				},
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newAdminTestHandlers(tc.create, nil, nil)

			r := httptest.NewRequest(http.MethodPost, "/restaurants", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			h.CreateRestaurantHandler(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body)
			}
			if tc.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tc.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tc.wantCode)
				}
			}
		})
	}
}

func TestCreateRestaurantHandler_DefaultsStatusToActive(t *testing.T) {
	var gotStatus string
	create := &fakeCreateRestaurant{
		fn: func(_ context.Context, input application.CreateRestaurantInput) (entities.Restaurant, error) {
			gotStatus = input.Status
			return mustRestaurant(), nil
		},
	}
	h := newAdminTestHandlers(create, nil, nil)

	body := `{"name":"Pizza Place","address":"Main St 123","phone_number":"+5511999999999","delivery_fee":5.0}`
	r := httptest.NewRequest(http.MethodPost, "/restaurants", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateRestaurantHandler(w, r)

	if gotStatus != "active" {
		t.Errorf("status = %q, want %q", gotStatus, "active")
	}
}

func TestCreateRestaurantHandler_PropagatesIdempotencyKey(t *testing.T) {
	var gotKey string
	create := &fakeCreateRestaurant{
		fn: func(_ context.Context, input application.CreateRestaurantInput) (entities.Restaurant, error) {
			gotKey = input.IdempotencyKey
			return mustRestaurant(), nil
		},
	}
	h := newAdminTestHandlers(create, nil, nil)

	body := `{"name":"Pizza Place","address":"Main St 123","phone_number":"+5511999999999","delivery_fee":5.0}`
	r := httptest.NewRequest(http.MethodPost, "/restaurants", strings.NewReader(body))
	r.Header.Set("Idempotency-Key", "client-key-1")
	w := httptest.NewRecorder()
	h.CreateRestaurantHandler(w, r)

	if gotKey != "client-key-1" {
		t.Errorf("idempotency key = %q, want %q", gotKey, "client-key-1")
	}
}

// ── CreateMenuItemHandler ───────────────────────────────────────────────────────

func TestCreateMenuItemHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		create     *fakeCreateMenuItem
		wantStatus int
		wantCode   string
	}{
		{
			name:       "success returns 201",
			body:       `{"name":"Margherita","price":39.9,"category":"pizzaria"}`,
			create:     &fakeCreateMenuItem{},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json returns 400",
			body:       `{bad json`,
			create:     &fakeCreateMenuItem{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name: "restaurant not found returns 404",
			body: `{"name":"Margherita","price":39.9,"category":"pizzaria"}`,
			create: &fakeCreateMenuItem{
				fn: func(_ context.Context, _ application.CreateMenuItemInput) (entities.MenuItem, error) {
					return entities.MenuItem{}, apperrors.NotFound("restaurant not found", nil)
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   apperrors.CodeNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newAdminTestHandlers(nil, tc.create, nil)

			r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/items", strings.NewReader(tc.body))
			r.SetPathValue("id", "rest-1")
			w := httptest.NewRecorder()
			h.CreateMenuItemHandler(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body)
			}
			if tc.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tc.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tc.wantCode)
				}
			}
		})
	}
}

func TestCreateMenuItemHandler_DefaultsAvailableToTrue(t *testing.T) {
	var gotAvailable bool
	create := &fakeCreateMenuItem{
		fn: func(_ context.Context, input application.CreateMenuItemInput) (entities.MenuItem, error) {
			gotAvailable = input.Available
			return mustMenuItem(), nil
		},
	}
	h := newAdminTestHandlers(nil, create, nil)

	body := `{"name":"Margherita","price":39.9,"category":"pizzaria"}`
	r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/items", strings.NewReader(body))
	r.SetPathValue("id", "rest-1")
	w := httptest.NewRecorder()
	h.CreateMenuItemHandler(w, r)

	if !gotAvailable {
		t.Error("expected available to default to true")
	}
}

// ── UpdateMenuItemHandler ───────────────────────────────────────────────────────

func TestUpdateMenuItemHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		update     *fakeUpdateMenuItem
		wantStatus int
		wantCode   string
	}{
		{
			name:       "success returns 200",
			body:       `{"available":false}`,
			update:     &fakeUpdateMenuItem{},
			wantStatus: http.StatusOK,
		},
		{
			name: "item not found returns 404",
			body: `{"available":false}`,
			update: &fakeUpdateMenuItem{
				fn: func(_ context.Context, _ application.UpdateMenuItemInput) (entities.MenuItem, error) {
					return entities.MenuItem{}, apperrors.NotFound("menu item not found", nil)
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   apperrors.CodeNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newAdminTestHandlers(nil, nil, tc.update)

			r := httptest.NewRequest(http.MethodPatch, "/restaurants/rest-1/menu/items/item-1", strings.NewReader(tc.body))
			r.SetPathValue("id", "rest-1")
			r.SetPathValue("itemId", "item-1")
			w := httptest.NewRecorder()
			h.UpdateMenuItemHandler(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body)
			}
			if tc.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tc.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tc.wantCode)
				}
			}
		})
	}
}

func TestUpdateMenuItemHandler_PropagatesPathParams(t *testing.T) {
	var gotRestaurantID, gotItemID string
	update := &fakeUpdateMenuItem{
		fn: func(_ context.Context, input application.UpdateMenuItemInput) (entities.MenuItem, error) {
			gotRestaurantID = input.RestaurantID
			gotItemID = input.ItemID
			return mustMenuItem(), nil
		},
	}
	h := newAdminTestHandlers(nil, nil, update)

	r := httptest.NewRequest(http.MethodPatch, "/restaurants/rest-1/menu/items/item-1", strings.NewReader(`{}`))
	r.SetPathValue("id", "rest-1")
	r.SetPathValue("itemId", "item-1")
	w := httptest.NewRecorder()
	h.UpdateMenuItemHandler(w, r)

	if gotRestaurantID != "rest-1" || gotItemID != "item-1" {
		t.Errorf("got restaurantID=%q itemID=%q, want rest-1/item-1", gotRestaurantID, gotItemID)
	}
}

// ── JSON round-trip of the DTOs ─────────────────────────────────────────────────

func TestMenuItemResponse_JSONShape(t *testing.T) {
	h := newAdminTestHandlers(nil, nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/items", strings.NewReader(`{"name":"Margherita","price":39.9,"category":"pizzaria"}`))
	r.SetPathValue("id", "rest-1")
	w := httptest.NewRecorder()
	h.CreateMenuItemHandler(w, r)

	var resp menuItemResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" || resp.RestaurantID == "" {
		t.Errorf("expected non-empty id and restaurant_id, got %+v", resp)
	}
}
