package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextKeysValues(t *testing.T) {
	if string(RequestIDKey) != "request_id" {
		t.Fatalf("unexpected request key: %s", RequestIDKey)
	}
	if string(CorrelationIDKey) != "correlation_id" {
		t.Fatalf("unexpected correlation key: %s", CorrelationIDKey)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Context().Value(RequestIDKey); got == nil || got.(string) == "" {
			t.Fatal("expected request id in context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	// generated
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	if rr.Header().Get(RequestIDHeader) == "" {
		t.Fatal("expected generated request id header")
	}

	// preserved
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "req-123")
	h.ServeHTTP(rr, req)
	if rr.Header().Get(RequestIDHeader) != "req-123" {
		t.Fatalf("expected request id to be preserved")
	}
}

func TestCorrelationIDMiddleware(t *testing.T) {
	h := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Context().Value(CorrelationIDKey); got == nil || got.(string) == "" {
			t.Fatal("expected correlation id in context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	// generated
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	if rr.Header().Get(CorrelationIDHeader) == "" {
		t.Fatal("expected generated correlation id header")
	}

	// preserved
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(CorrelationIDHeader, "corr-123")
	h.ServeHTTP(rr, req)
	if rr.Header().Get(CorrelationIDHeader) != "corr-123" {
		t.Fatalf("expected correlation id to be preserved")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	log := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	h := Recovery(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "internal server error") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestAccessLogMiddleware(t *testing.T) {
	buf := &bytes.Buffer{}
	log := slog.New(slog.NewJSONHandler(buf, nil))

	h := AccessLog(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req = req.WithContext(withIDs(req.Context()))
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	out := buf.String()
	for _, mustContain := range []string{"http_request", "request_id", "correlation_id", "/orders", "POST"} {
		if !strings.Contains(out, mustContain) {
			t.Fatalf("expected log to contain %q, got: %s", mustContain, out)
		}
	}
}

func withIDs(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, RequestIDKey, "req-1")
	ctx = context.WithValue(ctx, CorrelationIDKey, "corr-1")
	return ctx
}
