package httpdelivery

import (
	"net/http"

	apperrors "food_delivery_platform/shared/errors"
	"food_delivery_platform/shared/utils"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/observability"
)

type validateMenuItemsRequest struct {
	ItemIDs []string `json:"item_ids"`
}

type menuItemValidationResponse struct {
	ItemID    string  `json:"item_id"`
	Found     bool    `json:"found"`
	Available bool    `json:"available"`
	Price     float64 `json:"price"`
	Valid     bool    `json:"valid"`
	Reason    string  `json:"reason,omitempty"`
}

type validateMenuItemsResponse struct {
	RestaurantID     string                       `json:"restaurant_id"`
	RestaurantActive bool                         `json:"restaurant_active"`
	AllValid         bool                         `json:"all_valid"`
	Items            []menuItemValidationResponse `json:"items"`
}

// ValidateMenuItemsHandler e' um endpoint interno consumido pelo order-service para
// validar, antes de criar um pedido, se um restaurante esta ativo e se os itens
// pedidos existem/estao disponiveis, retornando o preco de catalogo de cada um.
// Devolve 200 mesmo quando itens sao invalidos: a invalidez e' informacao de
// negocio (RN003), nao um erro de transporte; so a ausencia do restaurante ou uma
// requisicao malformada geram erro HTTP.
func (h *RestaurantHandlers) ValidateMenuItemsHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.validateMenuItems == nil {
		writeAppError(w, r, apperrors.Internal("validate menu items use case is not configured", nil))
		return
	}

	var req validateMenuItemsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAppError(w, r, requestBodyError(err))
		return
	}

	out, err := h.validateMenuItems.Execute(r.Context(), application.ValidateMenuItemsInput{
		RestaurantID: r.PathValue("id"),
		ItemIDs:      req.ItemIDs,
	})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	recordValidation(r, out)
	_ = utils.WriteJSON(w, http.StatusOK, toValidateMenuItemsResponse(out))
}

func recordValidation(r *http.Request, out application.ValidateMenuItemsOutput) {
	observability.AnnotateMenuValidation(r.Context(), out.RestaurantID, len(out.Items), out.AllValid)
	for _, item := range out.Items {
		result := item.Reason
		if item.Valid {
			result = "valid"
		}
		observability.RecordMenuItemValidation(result)
	}
}

func toValidateMenuItemsResponse(out application.ValidateMenuItemsOutput) validateMenuItemsResponse {
	items := make([]menuItemValidationResponse, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, menuItemValidationResponse{
			ItemID:    item.ItemID,
			Found:     item.Found,
			Available: item.Available,
			Price:     item.Price,
			Valid:     item.Valid,
			Reason:    item.Reason,
		})
	}

	return validateMenuItemsResponse{
		RestaurantID:     out.RestaurantID,
		RestaurantActive: out.RestaurantActive,
		AllValid:         out.AllValid,
		Items:            items,
	}
}
