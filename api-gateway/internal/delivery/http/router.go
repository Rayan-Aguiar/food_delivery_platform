package httpdelivery

import (
	"log/slog"
	"net/http"
	"time"

	"food_delivery_platform/shared/middleware"
)

func NewRouter(log *slog.Logger, timeout time.Duration, authProxy http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", LiveHandler)
	mux.HandleFunc("GET /health/ready", ReadyHandler)

	RegisterAuthRoutes(mux, authProxy)

	mux.HandleFunc("/", notFoundHandler)

	var handler http.Handler = mux
	handler = middleware.Recovery(log)(handler)
	handler = middleware.AccessLog(log)(handler)
	handler = middleware.CorrelationID(handler)
	handler = middleware.RequestID(handler)
	handler = http.TimeoutHandler(handler, timeout, "request timeout")

	return handler
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "route not found", http.StatusNotFound)
}
