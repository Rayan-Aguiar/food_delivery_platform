package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// unmatchedRoute agrupa requests barradas antes do mux (ex.: method guard),
// mantendo a cardinalidade do label sob controle.
const unmatchedRoute = "unmatched"

var (
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duracao de requests HTTP em segundos.",
			Buckets: prometheus.DefBuckets,
		},
		// "path" recebe o padrao da rota (ex.: "GET /restaurants/{id}"), nunca o path
		// cru: com ids no path cada restaurante viraria uma serie nova no Prometheus.
		[]string{"method", "path", "status"},
	)

	menuItemValidationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "restaurant_menu_item_validations_total",
			Help: "Itens validados pelo endpoint interno, por resultado (valid ou motivo da invalidez).",
		},
		[]string{"result"},
	)
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// HTTPMetricsMiddleware precisa envolver o ServeMux diretamente: o mux preenche
// r.Pattern no mesmo *http.Request que recebe, e qualquer middleware que clone a
// request entre os dois (r.WithContext) esconderia o padrao da rota.
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		route := r.Pattern
		if route == "" {
			route = unmatchedRoute
		}

		httpRequestDuration.WithLabelValues(
			r.Method,
			route,
			strconv.Itoa(rw.statusCode),
		).Observe(time.Since(start).Seconds())
	})
}

// RecordMenuItemValidation conta um item validado. result deve ser "valid" ou um
// dos motivos fixos de invalidez, para o label ter cardinalidade limitada.
func RecordMenuItemValidation(result string) {
	menuItemValidationsTotal.WithLabelValues(result).Inc()
}
