package httpdelivery

import (
	"net/http"

	"food_delivery_platform/shared/contracts"
	apperrors "food_delivery_platform/shared/errors"
	"food_delivery_platform/shared/middleware"
	"food_delivery_platform/shared/utils"
)

func LiveHandler(w http.ResponseWriter, r *http.Request) {
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "live"})
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func AuthHealthHandler(w http.ResponseWriter, r *http.Request) {
	_ = utils.WriteJSON(w, http.StatusOK, map[string]string{"service": "auth-service", "status": "ok"})
}

func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	writeAppError(w, r, apperrors.InvalidArgument("method not allowed", nil))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeAppError(w, r, apperrors.NotFound("route not found", nil))
}

func writeAppError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := utils.StringFromContext(r.Context(), middleware.RequestIDKey)
	correlationID := utils.StringFromContext(r.Context(), middleware.CorrelationIDKey)
	status, response := apperrors.ToHTTPResponse(err, requestID, correlationID)
	_ = utils.WriteJSON(w, status, contracts.ErrorResponse{
		Code:          response.Code,
		Message:       response.Message,
		RequestID:     response.RequestID,
		CorrelationID: response.CorrelationID,
	})
}
