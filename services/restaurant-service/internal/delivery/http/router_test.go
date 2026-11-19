package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"food_delivery_platform/shared/contracts"
	apperrors "food_delivery_platform/shared/errors"
)

var silentLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func newTestRouter() http.Handler {
	restaurants := NewRestaurantHandlers(&fakeListRestaurants{}, &fakeGetRestaurantDetails{}, &fakeListRestaurantMenu{}, &fakeCreateRestaurant{}, &fakeCreateMenuItem{}, &fakeUpdateMenuItem{}, &fakeValidateMenuItems{})
	return NewRouter(silentLogger, 5*time.Second, "restaurant-service", restaurants, &fakeTokenValidator{})
}

func withBearerToken(r *http.Request) *http.Request {
	r.Header.Set("Authorization", "Bearer valid-token")
	return r
}

func TestRouter_ListRestaurants(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurants", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

func TestRouter_GetRestaurantDetails(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurants/rest-1", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

func TestRouter_ListRestaurantMenu(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurants/rest-1/menu", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

func TestRouter_CreateRestaurant(t *testing.T) {
	body := `{"name":"Pizza Place","address":"Main St 123","phone_number":"+5511999999999","delivery_fee":5.0}`
	r := withBearerToken(httptest.NewRequest(http.MethodPost, "/restaurants", strings.NewReader(body)))
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body)
	}
}

func TestRouter_CreateMenuItem(t *testing.T) {
	body := `{"name":"Margherita","price":39.9,"category":"pizzaria"}`
	r := withBearerToken(httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/items", strings.NewReader(body)))
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body)
	}
}

func TestRouter_UpdateMenuItem(t *testing.T) {
	body := `{"available":false}`
	r := withBearerToken(httptest.NewRequest(http.MethodPatch, "/restaurants/rest-1/menu/items/item-1", strings.NewReader(body)))
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

// ── JWT auth nas rotas administrativas ──────────────────────────────────────────

func TestRouter_AdminRoutes_RejectMissingToken(t *testing.T) {
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/restaurants", `{"name":"x","address":"y","phone_number":"z","delivery_fee":1}`},
		{http.MethodPost, "/restaurants/rest-1/menu/items", `{"name":"x","price":1,"category":"pizzaria"}`},
		{http.MethodPatch, "/restaurants/rest-1/menu/items/item-1", `{"available":false}`},
	}

	router := newTestRouter()
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401 for %s %s; body: %s", w.Code, tt.method, tt.path, w.Body)
			}
			var resp contracts.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Code != apperrors.CodeUnauthorized {
				t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeUnauthorized)
			}
		})
	}
}

func TestRouter_AdminRoutes_RejectInvalidToken(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/restaurants", strings.NewReader(`{"name":"x","address":"y","phone_number":"z","delivery_fee":1}`))
	r.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()

	restaurants := NewRestaurantHandlers(&fakeListRestaurants{}, &fakeGetRestaurantDetails{}, &fakeListRestaurantMenu{}, &fakeCreateRestaurant{}, &fakeCreateMenuItem{}, &fakeUpdateMenuItem{}, &fakeValidateMenuItems{})
	validator := &fakeTokenValidator{fn: func(_ context.Context, _ string) (string, error) {
		return "", errors.New("invalid signature")
	}}
	router := NewRouter(silentLogger, 5*time.Second, "restaurant-service", restaurants, validator)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body: %s", w.Code, w.Body)
	}
}

func TestRouter_ValidateMenuItems(t *testing.T) {
	body := `{"item_ids":["item-1"]}`
	r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/validate", strings.NewReader(body))
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

func TestRouter_NilRestaurantHandlers_HealthStillWorks(t *testing.T) {
	handler := NewRouter(silentLogger, 5*time.Second, "restaurant-service", nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestRouter_MethodGuard_RejectsWrongMethod(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/health/live"},
		{http.MethodPost, "/health/ready"},
	}

	router := newTestRouter()
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 for %s %s", w.Code, tt.method, tt.path)
			}
			var resp contracts.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Code != apperrors.CodeInvalidArgument {
				t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeInvalidArgument)
			}
		})
	}
}

func TestRouter_UnknownPath_Returns404(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/nonexistent", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Code != apperrors.CodeNotFound {
		t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeNotFound)
	}
}

func TestRouter_OversizedBody_Returns413(t *testing.T) {
	body := `{"item_ids":["` + strings.Repeat("x", maxRequestBodyBytes) + `"]}`
	r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/validate", strings.NewReader(body))
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413; body: %s", w.Code, w.Body)
	}
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Code != codePayloadTooLarge {
		t.Errorf("code = %q, want %q", resp.Code, codePayloadTooLarge)
	}
}

func TestRouter_MetricsEndpoint_ExposesRouteLabels(t *testing.T) {
	router := newTestRouter()

	// Gera uma observacao numa rota com path param antes de ler /metrics.
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/restaurants/rest-42", nil))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `path="GET /restaurants/{id}"`) {
		t.Error("expected histogram labeled by route pattern through the full middleware chain")
	}
	if strings.Contains(body, `path="/restaurants/rest-42"`) {
		t.Error("raw path with id must not be used as label")
	}
}
