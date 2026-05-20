package httpdelivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"food_delivery_platform/shared/contracts"
	apperrors "food_delivery_platform/shared/errors"
)

// ── LiveHandler ───────────────────────────────────────────────────────────────

func TestLiveHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	LiveHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "live" {
		t.Errorf("status = %q, want %q", body["status"], "live")
	}
}

// ── ReadyHandler ──────────────────────────────────────────────────────────────

func TestReadyHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	ReadyHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ready" {
		t.Errorf("status = %q, want %q", body["status"], "ready")
	}
}

// ── AuthHealthHandler ─────────────────────────────────────────────────────────

func TestAuthHealthHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/auth/health", nil)
	w := httptest.NewRecorder()

	AuthHealthHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["service"] != "auth-service" {
		t.Errorf("service = %q, want auth-service", body["service"])
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
}

// ── MethodNotAllowedHandler ───────────────────────────────────────────────────

func TestMethodNotAllowedHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	w := httptest.NewRecorder()

	MethodNotAllowedHandler(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode ErrorResponse: %v", err)
	}
	if resp.Code != apperrors.CodeInvalidArgument {
		t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeInvalidArgument)
	}
}

// ── NotFoundHandler ───────────────────────────────────────────────────────────

func TestNotFoundHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	NotFoundHandler(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode ErrorResponse: %v", err)
	}
	if resp.Code != apperrors.CodeNotFound {
		t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeNotFound)
	}
}

// ── writeAppError ─────────────────────────────────────────────────────────────

func TestWriteAppError_KnownAppError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	writeAppError(w, r, apperrors.Forbidden("acesso negado", nil))

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Code != apperrors.CodeForbidden {
		t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeForbidden)
	}
	if resp.Message != "acesso negado" {
		t.Errorf("message = %q, want %q", resp.Message, "acesso negado")
	}
}

func TestWriteAppError_UnknownError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	writeAppError(w, r, apperrors.Internal("falha inesperada", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Code != apperrors.CodeInternal {
		t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeInternal)
	}
}
