package httpdelivery

import (
	"log/slog"
	"net/http"
	"time"

	"food_delivery_platform/shared/middleware"
)

func NewRouter(log *slog.Logger, timeout time.Duration, serviceName string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", LiveHandler)
	mux.HandleFunc("GET /health/ready", ReadyHandler(serviceName))
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
	allowedMethodsByPath := map[string]string{
		"/health/live":  http.MethodGet,
		"/health/ready": http.MethodGet,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedMethod, ok := allowedMethodsByPath[r.URL.Path]; ok && r.Method != allowedMethod {
			MethodNotAllowedHandler(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
