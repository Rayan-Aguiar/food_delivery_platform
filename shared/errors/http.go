package apperrors

import (
	"errors"
	"net/http"

	"food_delivery_platform/shared/contracts"
)

func ToHTTPResponse(err error, requestID, correlationID string) (int, contracts.ErrorResponse) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode, contracts.ErrorResponse{
			Code:          appErr.Code,
			Message:       appErr.Message,
			RequestID:     requestID,
			CorrelationID: correlationID,
		}
	}

	return http.StatusInternalServerError, contracts.ErrorResponse{
		Code:          CodeInternal,
		Message:       "internal server error",
		RequestID:     requestID,
		CorrelationID: correlationID,
	}
}
