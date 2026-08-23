package httpdelivery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"food_delivery_platform/services/restaurant-service/internal/application"
	apperrors "food_delivery_platform/shared/errors"
)

func TestValidateMenuItemsHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		executor   *fakeValidateMenuItems
		wantStatus int
		wantCode   string
	}{
		{
			name:       "success returns 200",
			body:       `{"item_ids":["item-1"]}`,
			executor:   &fakeValidateMenuItems{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json returns 400",
			body:       `{bad json`,
			executor:   &fakeValidateMenuItems{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name: "empty item ids returns 400",
			body: `{"item_ids":[]}`,
			executor: &fakeValidateMenuItems{
				fn: func(_ context.Context, _ application.ValidateMenuItemsInput) (application.ValidateMenuItemsOutput, error) {
					return application.ValidateMenuItemsOutput{}, apperrors.InvalidArgument("at least one item id is required", nil)
				},
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name: "restaurant not found returns 404",
			body: `{"item_ids":["item-1"]}`,
			executor: &fakeValidateMenuItems{
				fn: func(_ context.Context, _ application.ValidateMenuItemsInput) (application.ValidateMenuItemsOutput, error) {
					return application.ValidateMenuItemsOutput{}, apperrors.NotFound("restaurant not found", nil)
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   apperrors.CodeNotFound,
		},
		{
			name: "unavailable item still returns 200 with valid=false",
			body: `{"item_ids":["item-1"]}`,
			executor: &fakeValidateMenuItems{
				fn: func(_ context.Context, input application.ValidateMenuItemsInput) (application.ValidateMenuItemsOutput, error) {
					return application.ValidateMenuItemsOutput{
						RestaurantID:     input.RestaurantID,
						RestaurantActive: true,
						AllValid:         false,
						Items: []application.MenuItemValidationResult{
							{ItemID: "item-1", Found: true, Available: false, Price: 39.9, Valid: false, Reason: application.ValidationReasonItemUnavailable},
						},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newAdminTestHandlers(nil, nil, nil)
			h.validateMenuItems = tc.executor

			r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/validate", strings.NewReader(tc.body))
			r.SetPathValue("id", "rest-1")
			w := httptest.NewRecorder()
			h.ValidateMenuItemsHandler(w, r)

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

func TestValidateMenuItemsHandler_NilExecutor(t *testing.T) {
	h := newAdminTestHandlers(nil, nil, nil)
	h.validateMenuItems = nil

	r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/validate", strings.NewReader(`{"item_ids":["item-1"]}`))
	r.SetPathValue("id", "rest-1")
	w := httptest.NewRecorder()
	h.ValidateMenuItemsHandler(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

func TestValidateMenuItemsHandler_ResponseShape(t *testing.T) {
	h := newAdminTestHandlers(nil, nil, nil)
	h.validateMenuItems = &fakeValidateMenuItems{
		fn: func(_ context.Context, input application.ValidateMenuItemsInput) (application.ValidateMenuItemsOutput, error) {
			return application.ValidateMenuItemsOutput{
				RestaurantID:     input.RestaurantID,
				RestaurantActive: true,
				AllValid:         false,
				Items: []application.MenuItemValidationResult{
					{ItemID: "item-1", Found: true, Available: true, Price: 39.9, Valid: true},
					{ItemID: "item-2", Found: false, Valid: false, Reason: application.ValidationReasonItemNotFound},
				},
			}, nil
		},
	}

	r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/validate", strings.NewReader(`{"item_ids":["item-1","item-2"]}`))
	r.SetPathValue("id", "rest-1")
	w := httptest.NewRecorder()
	h.ValidateMenuItemsHandler(w, r)

	var resp validateMenuItemsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.RestaurantID != "rest-1" {
		t.Errorf("restaurant_id = %q, want rest-1", resp.RestaurantID)
	}
	if !resp.RestaurantActive {
		t.Error("expected restaurant_active = true")
	}
	if resp.AllValid {
		t.Error("expected all_valid = false")
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].Price != 39.9 || !resp.Items[0].Valid {
		t.Errorf("unexpected item-1: %+v", resp.Items[0])
	}
	if resp.Items[1].Found || resp.Items[1].Reason != application.ValidationReasonItemNotFound {
		t.Errorf("unexpected item-2: %+v", resp.Items[1])
	}
}
