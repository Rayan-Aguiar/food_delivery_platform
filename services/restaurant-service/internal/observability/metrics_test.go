package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// seriesCount devolve quantas observacoes o histograma tem para o conjunto de labels.
func seriesCount(t *testing.T, labels prometheus.Labels) uint64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, mf := range families {
		if mf.GetName() != "http_request_duration_seconds" {
			continue
		}
		for _, m := range mf.GetMetric() {
			match := true
			for _, lp := range m.GetLabel() {
				if want, ok := labels[lp.GetName()]; ok && lp.GetValue() != want {
					match = false
				}
			}
			if match {
				return m.GetHistogram().GetSampleCount()
			}
		}
	}
	return 0
}

func TestHTTPMetricsMiddleware_LabelsByRoutePatternNotRawPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /things/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := HTTPMetricsMiddleware(mux)

	for _, id := range []string{"a", "b", "c"} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/things/"+id, nil))
	}

	if got := seriesCount(t, prometheus.Labels{"method": "GET", "path": "GET /things/{id}", "status": "418"}); got != 3 {
		t.Errorf("observations for route pattern = %d, want 3", got)
	}
	if got := seriesCount(t, prometheus.Labels{"path": "/things/a"}); got != 0 {
		t.Errorf("raw path must not become a label value, got %d observations", got)
	}
}

func TestHTTPMetricsMiddleware_UnmatchedRoute(t *testing.T) {
	// Handler que nao e' um ServeMux: r.Pattern fica vazio.
	handler := HTTPMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodDelete, "/whatever/123", nil))

	if got := seriesCount(t, prometheus.Labels{"method": "DELETE", "path": unmatchedRoute, "status": "405"}); got != 1 {
		t.Errorf("observations for unmatched route = %d, want 1", got)
	}
}

func TestRecordMenuItemValidation(t *testing.T) {
	before := testutil.ToFloat64(menuItemValidationsTotal.WithLabelValues("item_unavailable"))

	RecordMenuItemValidation("item_unavailable")
	RecordMenuItemValidation("item_unavailable")

	after := testutil.ToFloat64(menuItemValidationsTotal.WithLabelValues("item_unavailable"))
	if after-before != 2 {
		t.Errorf("counter delta = %v, want 2", after-before)
	}
}
