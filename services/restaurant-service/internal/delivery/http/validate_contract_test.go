package httpdelivery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"food_delivery_platform/services/restaurant-service/internal/application"
)

// Teste de contrato do endpoint interno POST /restaurants/{id}/menu/validate.
//
// O order-service depende deste JSON byte a byte (nomes de campo, tipos e a
// omissao de "reason" quando o item e' valido). Diferente dos testes de handler,
// que decodificam na propria struct de resposta, este compara com o JSON literal:
// renomear uma tag json ou trocar um tipo quebra o teste, como quebraria o consumidor.
// Mudou de proposito? Atualize tambem docs/openapi/openapi.v1.yaml.
func TestValidateMenuItemsContract(t *testing.T) {
	h := newAdminTestHandlers(nil, nil, nil)
	h.validateMenuItems = &fakeValidateMenuItems{
		fn: func(_ context.Context, in application.ValidateMenuItemsInput) (application.ValidateMenuItemsOutput, error) {
			return application.ValidateMenuItemsOutput{
				RestaurantID:     in.RestaurantID,
				RestaurantActive: true,
				AllValid:         false,
				Items: []application.MenuItemValidationResult{
					{ItemID: "item-1", Found: true, Available: true, Price: 39.9, Valid: true},
					{ItemID: "item-2", Found: true, Available: false, Price: 12.5, Valid: false, Reason: application.ValidationReasonItemUnavailable},
					{ItemID: "item-3", Found: false, Valid: false, Reason: application.ValidationReasonItemNotFound},
				},
			}, nil
		},
	}

	r := httptest.NewRequest(http.MethodPost, "/restaurants/rest-1/menu/validate",
		strings.NewReader(`{"item_ids":["item-1","item-2","item-3"]}`))
	r.SetPathValue("id", "rest-1")
	w := httptest.NewRecorder()
	h.ValidateMenuItemsHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (itens invalidos sao resposta de negocio, nao erro HTTP)", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	const want = `{
		"restaurant_id": "rest-1",
		"restaurant_active": true,
		"all_valid": false,
		"items": [
			{"item_id": "item-1", "found": true,  "available": true,  "price": 39.9, "valid": true},
			{"item_id": "item-2", "found": true,  "available": false, "price": 12.5, "valid": false, "reason": "item_unavailable"},
			{"item_id": "item-3", "found": false, "available": false, "price": 0,    "valid": false, "reason": "item_not_found"}
		]
	}`

	assertJSONEqual(t, w.Body.Bytes(), []byte(want))
}

// TestValidateMenuItemsContract_ReasonValues fixa os motivos que o consumidor
// pode receber. Adicionar um motivo novo e' mudanca de contrato.
func TestValidateMenuItemsContract_ReasonValues(t *testing.T) {
	got := []string{
		application.ValidationReasonItemNotFound,
		application.ValidationReasonItemUnavailable,
		application.ValidationReasonRestaurantInactive,
	}
	want := []string{"item_not_found", "item_unavailable", "restaurant_inactive"}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("reason[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func assertJSONEqual(t *testing.T, got, want []byte) {
	t.Helper()

	var gotV, wantV any
	if err := json.Unmarshal(got, &gotV); err != nil {
		t.Fatalf("response is not valid JSON: %v\n%s", err, got)
	}
	if err := json.Unmarshal(want, &wantV); err != nil {
		t.Fatalf("expected value is not valid JSON: %v", err)
	}

	gotNorm, _ := json.Marshal(gotV)
	wantNorm, _ := json.Marshal(wantV)
	if string(gotNorm) != string(wantNorm) {
		t.Errorf("contract mismatch\n got:  %s\n want: %s", gotNorm, wantNorm)
	}
}
