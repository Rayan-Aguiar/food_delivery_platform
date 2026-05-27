package proxy

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"food_delivery_platform/shared/contracts"
	"food_delivery_platform/shared/middleware"
	"food_delivery_platform/shared/utils"
)

func NewAuthProxy(authServiceURL string, log *slog.Logger) (http.Handler, error) {
	target, err := url.Parse(authServiceURL)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(target)
	originalDirector := rp.Director

	rp.Director = func(req *http.Request) {
		originalDirector(req)

		reqID, _ := req.Context().Value(middleware.RequestIDKey).(string)
		corrID, _ := req.Context().Value(middleware.CorrelationIDKey).(string)

		if reqID != "" {
			req.Header.Set(middleware.RequestIDHeader, reqID)
		}
		if corrID != "" {
			req.Header.Set(middleware.CorrelationIDHeader, corrID)
		}

	}

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		status := http.StatusBadGateway
		message := "upstream auth unavailable"

		if err == context.DeadlineExceeded {
			status = http.StatusGatewayTimeout
			message = "upstream auth timeout"
		}

		reqID, _ := r.Context().Value(middleware.RequestIDKey).(string)
		corrID, _ := r.Context().Value(middleware.CorrelationIDKey).(string)

		log.Error("auth upstream error",
			"error", err.Error(),
			"path", r.URL.Path,
			"request_id", reqID,
			"correlation_id", corrID,
		)

		_ = utils.WriteJSON(w, status, contracts.ErrorResponse{
			Code:          "UPSTREAM_ERROR",
			Message:       message,
			RequestID:     reqID,
			CorrelationID: corrID,
		})
	}
	return rp, nil
}
