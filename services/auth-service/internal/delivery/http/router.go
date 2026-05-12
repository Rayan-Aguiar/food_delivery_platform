package httpdelivery

import (
	"log/slog"
	"net/http"
	"time"

	"food_delivery_platform/shared/middleware"
)

func NewRouter(log *slog.Logger, timeout time.Duration) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", LiveHandler)
	mux.HandleFunc("GET /health/ready", ReadyHandler)
	mux.HandleFunc("GET /auth/health", AuthHealthHandler)

	mux.HandleFunc("/", NotFoundHandler)

	var handler http.Handler = mux
	handler = methodGuard(handler)
	handler = middleware.Recovery(log)(handler)
	handler = middleware.AccessLog(log)(handler)
	handler = middleware.CorrelationID(handler)
	handler = middleware.RequestID(handler)
	handler = http.TimeoutHandler(handler, timeout, "request timeout")

	return handler
}

func methodGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health/live" || r.URL.Path == "/health/ready" || r.URL.Path == "/auth/health" {
			if r.Method != http.MethodGet {
				MethodNotAllowedHandler(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
