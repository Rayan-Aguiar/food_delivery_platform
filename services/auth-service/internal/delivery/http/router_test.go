package httpdelivery

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"food_delivery_platform/shared/contracts"
	apperrors "food_delivery_platform/shared/errors"
)

// silentLogger descarta toda saída de log para não poluir a saída dos testes.
var silentLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// newTestRouter cria um router com auth handlers usando fakes padrão (sucesso).
func newTestRouter() http.Handler {
	auth := NewAuthHandlers(&fakeRegister{}, &fakeLogin{}, &fakeRefresh{}, &fakeLogout{})
	return NewRouter(silentLogger, 5*time.Second, auth)
}

// ── Rotas de health ───────────────────────────────────────────────────────────

func TestRouter_HealthLive(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestRouter_HealthReady(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestRouter_AuthHealth(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/auth/health", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// ── Rotas de autenticação ─────────────────────────────────────────────────────

func TestRouter_Register(t *testing.T) {
	body := `{"email":"user@example.com","password":"pw"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body)
	}
}

func TestRouter_Login(t *testing.T) {
	body := `{"email":"user@example.com","password":"pw"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	r.RemoteAddr = "10.0.0.1:9999"
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

func TestRouter_Refresh(t *testing.T) {
	body := `{"refresh_token":"some-token"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
	r.RemoteAddr = "10.0.0.2:9999"
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

func TestRouter_Logout(t *testing.T) {
	body := `{"refresh_token":"some-token"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(body))
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", w.Code, w.Body)
	}
}

// ── Sem auth handlers (nil) ───────────────────────────────────────────────────

func TestRouter_NilAuthHandlers_HealthStillWorks(t *testing.T) {
	handler := NewRouter(silentLogger, 5*time.Second, nil)

	r := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// ── Method guard ──────────────────────────────────────────────────────────────

func TestRouter_MethodGuard_RejectsWrongMethod(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/health/live"},
		{http.MethodPost, "/health/ready"},
		{http.MethodPost, "/auth/health"},
		{http.MethodGet, "/auth/register"},
		{http.MethodGet, "/auth/login"},
		{http.MethodGet, "/auth/refresh"},
		{http.MethodGet, "/auth/logout"},
	}

	router := newTestRouter()
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 para %s %s", w.Code, tt.method, tt.path)
			}
			var resp contracts.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Code != apperrors.CodeInvalidArgument {
				t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeInvalidArgument)
			}
		})
	}
}

// ── Rota desconhecida ─────────────────────────────────────────────────────────

func TestRouter_UnknownPath_Returns404(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/nonexistent", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Code != apperrors.CodeNotFound {
		t.Errorf("code = %q, want %q", resp.Code, apperrors.CodeNotFound)
	}
}

// ── Middleware: RequestID e CorrelationID injetados ───────────────────────────

func TestRouter_RequestIDPropagatedInErrorResponse(t *testing.T) {
	// Qualquer rota desconhecida → 404 com request_id preenchido pelo middleware
	r := httptest.NewRequest(http.MethodGet, "/nope", nil)
	w := httptest.NewRecorder()
	newTestRouter().ServeHTTP(w, r)

	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.RequestID == "" {
		t.Error("request_id deve ser preenchido pelo middleware RequestID")
	}
}

func TestRouter_CORSPreflight(t *testing.T) {
	r := httptest.NewRequest(http.MethodOptions, "/auth/login", nil)
	r.Header.Set("Origin", "http://localhost:8085")
	r.Header.Set("Access-Control-Request-Method", http.MethodPost)
	w := httptest.NewRecorder()

	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow-origin = %q, want *", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected Access-Control-Allow-Methods header")
	}
}

func TestRouter_CORSHeadersOnRegularRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	r.Header.Set("Origin", "http://localhost:8085")
	w := httptest.NewRecorder()

	newTestRouter().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow-origin = %q, want *", got)
	}
	if got := w.Header().Get("Access-Control-Expose-Headers"); got == "" {
		t.Fatal("expected Access-Control-Expose-Headers header")
	}
}
