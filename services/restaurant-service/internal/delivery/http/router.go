package httpdelivery

import (
	"log/slog"
	"net/http"
	"time"

	"food_delivery_platform/shared/middleware"
)

func NewRouter(log *slog.Logger, timeout time.Duration, serviceName string, restaurants *RestaurantHandlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", LiveHandler)
	mux.HandleFunc("GET /health/ready", ReadyHandler(serviceName))

	if restaurants != nil {
		mux.HandleFunc("GET /restaurants", restaurants.ListRestaurantsHandler)
		mux.HandleFunc("GET /restaurants/{id}", restaurants.GetRestaurantDetailsHandler)
		mux.HandleFunc("GET /restaurants/{id}/menu", restaurants.ListRestaurantMenuHandler)

		// Rotas administrativas: protecao por JWT ainda nao aplicada (ver Fase 7).
		mux.HandleFunc("POST /restaurants", restaurants.CreateRestaurantHandler)
		mux.HandleFunc("POST /restaurants/{id}/menu/items", restaurants.CreateMenuItemHandler)
		mux.HandleFunc("PATCH /restaurants/{id}/menu/items/{itemId}", restaurants.UpdateMenuItemHandler)
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
