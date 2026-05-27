package httpdelivery

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"food_delivery_platform/api-gateway/internal/proxy"
	"food_delivery_platform/shared/middleware"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewRouterHealthAndNotFound(t *testing.T) {
	router := NewRouter(testLogger(), 2*time.Second, nil)

	t.Run("live", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		if rr.Header().Get(middleware.RequestIDHeader) == "" {
			t.Fatal("expected X-Request-ID response header")
		}
		if rr.Header().Get(middleware.CorrelationIDHeader) == "" {
			t.Fatal("expected X-Correlation-ID response header")
		}
	})

	t.Run("ready", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})
}

func TestRouterAuthRoutesAndHeaderPropagation(t *testing.T) {
	type upstreamCapture struct {
		Method        string
		Path          string
		Query         string
		Body          string
		RequestID     string
		CorrelationID string
	}

	captureCh := make(chan upstreamCapture, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := io.ReadAll(r.Body)
		captureCh <- upstreamCapture{
			Method:        r.Method,
			Path:          r.URL.Path,
			Query:         r.URL.RawQuery,
			Body:          string(payload),
			RequestID:     r.Header.Get(middleware.RequestIDHeader),
			CorrelationID: r.Header.Get(middleware.CorrelationIDHeader),
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	authProxy, err := proxy.NewAuthProxy(upstream.URL, testLogger())
	if err != nil {
		t.Fatalf("unexpected error creating auth proxy: %v", err)
	}

	router := NewRouter(testLogger(), 3*time.Second, authProxy)

	body := []byte(`{"email":"user@example.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login?source=gateway", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.RequestIDHeader, "req-abc")
	req.Header.Set(middleware.CorrelationIDHeader, "corr-xyz")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 from upstream passthrough, got %d", rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != `{"ok":true}` {
		t.Fatalf("expected passthrough body {\"ok\":true}, got %s", rr.Body.String())
	}

	select {
	case got := <-captureCh:
		if got.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", got.Method)
		}
		if got.Path != "/auth/login" {
			t.Fatalf("expected path /auth/login, got %s", got.Path)
		}
		if got.Query != "source=gateway" {
			t.Fatalf("expected query source=gateway, got %s", got.Query)
		}
		if got.Body != string(body) {
			t.Fatalf("expected body %s, got %s", string(body), got.Body)
		}
		if got.RequestID != "req-abc" {
			t.Fatalf("expected request id req-abc, got %s", got.RequestID)
		}
		if got.CorrelationID != "corr-xyz" {
			t.Fatalf("expected correlation id corr-xyz, got %s", got.CorrelationID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting upstream capture")
	}
}

func TestRouterMapsAllAuthEndpoints(t *testing.T) {
	var mu sync.Mutex
	hits := map[string]int{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hits[r.Method+" "+r.URL.Path]++
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	authProxy, err := proxy.NewAuthProxy(upstream.URL, testLogger())
	if err != nil {
		t.Fatalf("unexpected error creating auth proxy: %v", err)
	}

	router := NewRouter(testLogger(), 2*time.Second, authProxy)

	endpoints := []string{"/auth/register", "/auth/login", "/auth/refresh", "/auth/logout"}
	for _, endpoint := range endpoints {
		req := httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader([]byte(`{}`)))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected 204 for %s, got %d", endpoint, rr.Code)
		}
	}

	mu.Lock()
	defer mu.Unlock()

	for _, endpoint := range endpoints {
		key := http.MethodPost + " " + endpoint
		if hits[key] != 1 {
			t.Fatalf("expected one upstream hit for %s, got %d", key, hits[key])
		}
	}
}

func TestRegisterAuthRoutesNilProxyDoesNotExposeAuthEndpoints(t *testing.T) {
	router := NewRouter(testLogger(), 2*time.Second, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when auth proxy is nil, got %d", rr.Code)
	}
}

func TestHealthHandlerResponseShape(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rr := httptest.NewRecorder()

	LiveHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", payload["status"])
	}
	if payload["service"] != "api-gateway" {
		t.Fatalf("expected service=api-gateway, got %v", payload["service"])
	}
}
