package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	loginAttemptsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_attempts_total",
		Help: "Total de tentativas de login.",
	})

	loginFailuresTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_failures_total",
		Help: "Total de falhas de login.",
	})

	tokenRefreshTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_token_refresh_total",
		Help: "Total de tentativas de refresh token.",
	})

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duracao de requests HTTP em segundos.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
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

func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		httpRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(rw.statusCode),
		).Observe(time.Since(start).Seconds())
	})
}

func IncLoginAttempt() {
	loginAttemptsTotal.Inc()
}

func IncLoginFailure() {
	loginFailuresTotal.Inc()
}

func IncTokenRefresh() {
	tokenRefreshTotal.Inc()
}
