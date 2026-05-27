package proxy

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"food_delivery_platform/shared/contracts"
	"food_delivery_platform/shared/middleware"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewAuthProxyInvalidURL(t *testing.T) {
	proxy, err := NewAuthProxy("://invalid-url", testLogger())
	if err == nil {
		t.Fatal("expected error for invalid auth service URL, got nil")
	}
	if proxy != nil {
		t.Fatal("expected nil proxy when URL is invalid")
	}
}

func TestAuthProxyUpstreamUnavailableReturnsStandardError(t *testing.T) {
	proxy, err := NewAuthProxy("http://127.0.0.1:1", testLogger())
	if err != nil {
		t.Fatalf("unexpected error creating proxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req-1")
	ctx = context.WithValue(ctx, middleware.CorrelationIDKey, "corr-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rr.Code)
	}

	var resp contracts.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode json response: %v", err)
	}

	if resp.Code != "UPSTREAM_ERROR" {
		t.Fatalf("expected code UPSTREAM_ERROR, got %q", resp.Code)
	}
	if resp.Message == "" {
		t.Fatal("expected non-empty error message")
	}
	if resp.RequestID != "req-1" {
		t.Fatalf("expected request_id req-1, got %q", resp.RequestID)
	}
	if resp.CorrelationID != "corr-1" {
		t.Fatalf("expected correlation_id corr-1, got %q", resp.CorrelationID)
	}
}
