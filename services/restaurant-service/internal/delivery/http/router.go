package httpdelivery

import (
	"log/slog"
	"net/http"
	"time"

	"food_delivery_platform/services/restaurant-service/internal/observability"
	"food_delivery_platform/shared/middleware"

	"golang.org/x/time/rate"
)

func NewRouter(log *slog.Logger, timeout time.Duration, serviceName string, restaurants *RestaurantHandlers, tokens tokenValidator) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", LiveHandler)
	mux.HandleFunc("GET /health/ready", ReadyHandler(serviceName))
	mux.Handle("GET /metrics", observability.PrometheusHandler())

	if restaurants != nil {
		// Leitura publica: rate limit basico por IP contra burst/scraping.
		publicRL := newIPRateLimiter(rate.Limit(10), 30)
		mux.Handle("GET /restaurants", publicRL.Middleware(http.HandlerFunc(restaurants.ListRestaurantsHandler)))
		mux.Handle("GET /restaurants/{id}", publicRL.Middleware(http.HandlerFunc(restaurants.GetRestaurantDetailsHandler)))
		mux.Handle("GET /restaurants/{id}/menu", publicRL.Middleware(http.HandlerFunc(restaurants.ListRestaurantMenuHandler)))

		// Rotas administrativas: exigem Bearer token valido emitido pelo auth-service.
		adminAuth := JWTAuth(tokens)
		mux.Handle("POST /restaurants", adminAuth(http.HandlerFunc(restaurants.CreateRestaurantHandler)))
		mux.Handle("POST /restaurants/{id}/menu/items", adminAuth(http.HandlerFunc(restaurants.CreateMenuItemHandler)))
		mux.Handle("PATCH /restaurants/{id}/menu/items/{itemId}", adminAuth(http.HandlerFunc(restaurants.UpdateMenuItemHandler)))

		// Rota interna consumida pelo order-service: autenticacao servico-a-servico
		// (fora do escopo de JWT de usuario) fica para uma fase futura.
		mux.HandleFunc("POST /restaurants/{id}/menu/validate", restaurants.ValidateMenuItemsHandler)
	}

	mux.HandleFunc("/", NotFoundHandler)

	// Metricas e tracing ficam colados no mux (sem middleware que clone a request
	// no meio) para enxergar r.Pattern e rotular pela rota, nao pelo path cru.
	var handler http.Handler = mux
	handler = observability.HTTPMetricsMiddleware(handler)
	handler = observability.NewHTTPHandler(handler, "restaurant.http")
	handler = methodGuard(handler)
	handler = http.MaxBytesHandler(handler, maxRequestBodyBytes)
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
		"/metrics":      http.MethodGet,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedMethod, ok := allowedMethodsByPath[r.URL.Path]; ok && r.Method != allowedMethod {
			MethodNotAllowedHandler(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
