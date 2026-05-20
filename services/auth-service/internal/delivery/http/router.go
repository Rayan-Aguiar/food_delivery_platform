package httpdelivery

import (
	"log/slog"
	"net/http"
	"time"

	"food_delivery_platform/shared/middleware"

	"golang.org/x/time/rate"
)

func NewRouter(log *slog.Logger, timeout time.Duration, auth *AuthHandlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", LiveHandler)
	mux.HandleFunc("GET /health/ready", ReadyHandler)
	mux.HandleFunc("GET /auth/health", AuthHealthHandler)
	if auth != nil {
		// 5 requisições por minuto por IP, burst de 5 — proteção contra brute-force
		rl := newIPRateLimiter(rate.Every(12*time.Second), 5)

		mux.HandleFunc("POST /auth/register", auth.RegisterHandler)
		mux.Handle("POST /auth/login", rl.Middleware(http.HandlerFunc(auth.LoginHandler)))
		mux.Handle("POST /auth/refresh", rl.Middleware(http.HandlerFunc(auth.RefreshHandler)))
		mux.HandleFunc("POST /auth/logout", auth.LogoutHandler)
	}

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
		"/health/live":   http.MethodGet,
		"/health/ready":  http.MethodGet,
		"/auth/health":   http.MethodGet,
		"/auth/register": http.MethodPost,
		"/auth/login":    http.MethodPost,
		"/auth/refresh":  http.MethodPost,
		"/auth/logout":   http.MethodPost,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedMethod, ok := allowedMethodsByPath[r.URL.Path]; ok && r.Method != allowedMethod {
			MethodNotAllowedHandler(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
