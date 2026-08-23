package httpdelivery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	"food_delivery_platform/shared/contracts"
	apperrors "food_delivery_platform/shared/errors"
)

// ── ListRestaurantsHandler ──────────────────────────────────────────────────────

func TestListRestaurantsHandler(t *testing.T) {
	tests := []struct {
		name       string
		executor   *fakeListRestaurants
		wantStatus int
		wantCode   string
		wantCount  int
	}{
		{
			name:       "success returns restaurants",
			executor:   &fakeListRestaurants{},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name: "empty list returns empty array",
			executor: &fakeListRestaurants{
				fn: func(_ context.Context) ([]entities.Restaurant, error) {
					return []entities.Restaurant{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "use case error returns 500",
			executor: &fakeListRestaurants{
				fn: func(_ context.Context) ([]entities.Restaurant, error) {
					return nil, apperrors.Internal("db unavailable", nil)
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewRestaurantHandlers(tc.executor, &fakeGetRestaurantDetails{}, &fakeListRestaurantMenu{}, &fakeCreateRestaurant{}, &fakeCreateMenuItem{}, &fakeUpdateMenuItem{})

			r := httptest.NewRequest(http.MethodGet, "/restaurants", nil)
			w := httptest.NewRecorder()
			h.ListRestaurantsHandler(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body)
			}

			if tc.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tc.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tc.wantCode)
				}
				return
			}

			var resp listRestaurantsResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(resp.Restaurants) != tc.wantCount {
				t.Errorf("restaurants count = %d, want %d", len(resp.Restaurants), tc.wantCount)
			}
		})
	}
}

func TestListRestaurantsHandler_NilExecutor(t *testing.T) {
	h := NewRestaurantHandlers(nil, &fakeGetRestaurantDetails{}, &fakeListRestaurantMenu{}, &fakeCreateRestaurant{}, &fakeCreateMenuItem{}, &fakeUpdateMenuItem{})

	r := httptest.NewRequest(http.MethodGet, "/restaurants", nil)
	w := httptest.NewRecorder()
	h.ListRestaurantsHandler(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

// ── GetRestaurantDetailsHandler ─────────────────────────────────────────────────

func TestGetRestaurantDetailsHandler(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		executor   *fakeGetRestaurantDetails
		wantStatus int
		wantCode   string
	}{
		{
			name:       "success returns restaurant",
			id:         "rest-1",
			executor:   &fakeGetRestaurantDetails{},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found returns 404",
			id:   "missing",
			executor: &fakeGetRestaurantDetails{
				fn: func(_ context.Context, _ application.GetRestaurantDetailsInput) (entities.Restaurant, error) {
					return entities.Restaurant{}, apperrors.NotFound("restaurant not found", nil)
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   apperrors.CodeNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewRestaurantHandlers(&fakeListRestaurants{}, tc.executor, &fakeListRestaurantMenu{}, &fakeCreateRestaurant{}, &fakeCreateMenuItem{}, &fakeUpdateMenuItem{})

			r := httptest.NewRequest(http.MethodGet, "/restaurants/"+tc.id, nil)
			r.SetPathValue("id", tc.id)
			w := httptest.NewRecorder()
			h.GetRestaurantDetailsHandler(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body)
			}

			if tc.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tc.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tc.wantCode)
				}
				return
			}

			var resp restaurantResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.ID != "rest-1" {
				t.Errorf("id = %q, want rest-1", resp.ID)
			}
		})
	}
}

// ── ListRestaurantMenuHandler ───────────────────────────────────────────────────

func TestListRestaurantMenuHandler(t *testing.T) {
	tests := []struct {
		name       string
		executor   *fakeListRestaurantMenu
		wantStatus int
		wantCode   string
		wantCount  int
	}{
		{
			name:       "success returns menu items",
			executor:   &fakeListRestaurantMenu{},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name: "restaurant not found returns 404",
			executor: &fakeListRestaurantMenu{
				fn: func(_ context.Context, _ application.ListRestaurantMenuInput) ([]entities.MenuItem, error) {
					return nil, apperrors.NotFound("restaurant not found", nil)
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   apperrors.CodeNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewRestaurantHandlers(&fakeListRestaurants{}, &fakeGetRestaurantDetails{}, tc.executor, &fakeCreateRestaurant{}, &fakeCreateMenuItem{}, &fakeUpdateMenuItem{})

			r := httptest.NewRequest(http.MethodGet, "/restaurants/rest-1/menu", nil)
			r.SetPathValue("id", "rest-1")
			w := httptest.NewRecorder()
			h.ListRestaurantMenuHandler(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body)
			}

			if tc.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tc.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tc.wantCode)
				}
				return
			}

			var resp listRestaurantMenuResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(resp.Items) != tc.wantCount {
				t.Errorf("items count = %d, want %d", len(resp.Items), tc.wantCount)
			}
		})
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────────

func decodeError(t *testing.T, body *httptest.ResponseRecorder) contracts.ErrorResponse {
	t.Helper()
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(body.Body).Decode(&resp); err != nil {
		t.Fatalf("decode ErrorResponse: %v", err)
	}
	return resp
}
