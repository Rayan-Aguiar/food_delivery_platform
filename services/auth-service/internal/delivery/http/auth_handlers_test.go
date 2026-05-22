package httpdelivery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"food_delivery_platform/services/auth-service/internal/application"
	"food_delivery_platform/shared/contracts"
	apperrors "food_delivery_platform/shared/errors"
)

// newHandlers é um atalho para criar AuthHandlers com fakes padrão,
// substituindo apenas o executor indicado pelo chamador.
func newHandlers(reg registerUserExecutor, log loginUserExecutor, ref refreshTokenExecutor, out logoutSessionExecutor) *AuthHandlers {
	if reg == nil {
		reg = &fakeRegister{}
	}
	if log == nil {
		log = &fakeLogin{}
	}
	if ref == nil {
		ref = &fakeRefresh{}
	}
	if out == nil {
		out = &fakeLogout{}
	}
	return NewAuthHandlers(reg, log, ref, out)
}

// decodeError decodifica um ErrorResponse do corpo da resposta.
func decodeError(t *testing.T, body *httptest.ResponseRecorder) contracts.ErrorResponse {
	t.Helper()
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(body.Body).Decode(&resp); err != nil {
		t.Fatalf("decodificar ErrorResponse: %v", err)
	}
	return resp
}

// ── RegisterHandler ───────────────────────────────────────────────────────────

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		register   *fakeRegister
		wantStatus int
		wantCode   string
	}{
		{
			name:       "sucesso retorna 201 com ids",
			body:       `{"email":"user@example.com","password":"Valid1!"}`,
			register:   &fakeRegister{},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "json invalido retorna 400",
			body:       `{bad json`,
			register:   &fakeRegister{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name:       "campo desconhecido retorna 400",
			body:       `{"email":"user@example.com","password":"Valid1!","extra":"x"}`,
			register:   &fakeRegister{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name:       "executor nao configurado retorna 500",
			body:       `{"email":"user@example.com","password":"Valid1!"}`,
			register:   nil, // passa nil para simular use case não wired
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperrors.CodeInternal,
		},
		{
			name: "conflito de email retorna 409",
			body: `{"email":"user@example.com","password":"Valid1!"}`,
			register: &fakeRegister{fn: func(_ context.Context, _ application.RegisterUserInput) (application.RegisterUserOutput, error) {
				return application.RegisterUserOutput{}, apperrors.Conflict("email already registered", nil)
			}},
			wantStatus: http.StatusConflict,
			wantCode:   apperrors.CodeConflict,
		},
		{
			name: "erro interno do use case retorna 500",
			body: `{"email":"user@example.com","password":"Valid1!"}`,
			register: &fakeRegister{fn: func(_ context.Context, _ application.RegisterUserInput) (application.RegisterUserOutput, error) {
				return application.RegisterUserOutput{}, apperrors.Internal("db failure", nil)
			}},
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperrors.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reg registerUserExecutor = &fakeRegister{}
			if tt.register != nil {
				reg = tt.register
			} else {
				reg = nil
			}
			h := NewAuthHandlers(reg, &fakeLogin{}, &fakeRefresh{}, &fakeLogout{})
			r := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.RegisterHandler(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body)
			}
			if tt.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestRegisterHandler_SuccessBody(t *testing.T) {
	h := newHandlers(nil, nil, nil, nil)
	r := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"email":"u@e.com","password":"pw"}`))
	w := httptest.NewRecorder()

	h.RegisterHandler(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	var resp registerResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode registerResponse: %v", err)
	}
	if resp.UserID != "user-1" {
		t.Errorf("user_id = %q, want %q", resp.UserID, "user-1")
	}
	if resp.CredentialID != "cred-1" {
		t.Errorf("credential_id = %q, want %q", resp.CredentialID, "cred-1")
	}
	if resp.AccessToken != defaultTokens.AccessToken {
		t.Errorf("access_token = %q, want %q", resp.AccessToken, defaultTokens.AccessToken)
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("token_type = %q, want Bearer", resp.TokenType)
	}
}

// ── LoginHandler ──────────────────────────────────────────────────────────────

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		login      *fakeLogin
		wantStatus int
		wantCode   string
	}{
		{
			name:       "sucesso retorna 200 com tokens",
			body:       `{"email":"user@example.com","password":"pw"}`,
			login:      &fakeLogin{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "json invalido retorna 400",
			body:       `not json`,
			login:      &fakeLogin{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name: "credenciais invalidas retorna 401",
			body: `{"email":"user@example.com","password":"wrong"}`,
			login: &fakeLogin{fn: func(_ context.Context, _ application.LoginUserInput) (application.LoginUserOutput, error) {
				return application.LoginUserOutput{}, apperrors.Unauthorized("invalid credentials", nil)
			}},
			wantStatus: http.StatusUnauthorized,
			wantCode:   apperrors.CodeUnauthorized,
		},
		{
			name:       "executor nao configurado retorna 500",
			body:       `{"email":"user@example.com","password":"pw"}`,
			login:      nil,
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperrors.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log loginUserExecutor
			if tt.login != nil {
				log = tt.login
			}
			h := NewAuthHandlers(&fakeRegister{}, log, &fakeRefresh{}, &fakeLogout{})
			r := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			r.RemoteAddr = "10.0.0.1:4321"
			w := httptest.NewRecorder()

			h.LoginHandler(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body)
			}
			if tt.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestLoginHandler_SuccessBody(t *testing.T) {
	h := newHandlers(nil, nil, nil, nil)
	r := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"u@e.com","password":"pw"}`))
	r.RemoteAddr = "10.0.0.1:4321"
	w := httptest.NewRecorder()

	h.LoginHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp tokenResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode tokenResponse: %v", err)
	}
	if resp.AccessToken != defaultTokens.AccessToken {
		t.Errorf("access_token = %q, want %q", resp.AccessToken, defaultTokens.AccessToken)
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("token_type = %q, want Bearer", resp.TokenType)
	}
}

func TestRegisterHandler_PropagatesContextMetadata(t *testing.T) {
	var got application.RegisterUserInput
	h := NewAuthHandlers(&fakeRegister{fn: func(_ context.Context, in application.RegisterUserInput) (application.RegisterUserOutput, error) {
		got = in
		return application.RegisterUserOutput{UserID: "user-1", CredentialID: "cred-1", Tokens: defaultTokens}, nil
	}}, &fakeLogin{}, &fakeRefresh{}, &fakeLogout{})

	r := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"email":"user@example.com","password":"Valid1!"}`))
	r.Header.Set("X-Correlation-ID", "corr-123")
	r.Header.Set("X-Request-ID", "req-123")
	r.Header.Set("traceparent", "00-aabbccddeeff00112233445566778899-0011223344556677-01")
	r.Header.Set("Idempotency-Key", "idem-123")
	r.RemoteAddr = "10.0.0.1:5000"

	w := httptest.NewRecorder()
	httpHandler := NewRouter(silentLogger, 5*time.Second, h)
	httpHandler.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if got.CorrelationID != "corr-123" {
		t.Errorf("correlation id = %q, want %q", got.CorrelationID, "corr-123")
	}
	if got.CausationID != "req-123" {
		t.Errorf("causation id = %q, want %q", got.CausationID, "req-123")
	}
	if got.Traceparent == "" || got.IdempotencyKey != "idem-123" {
		t.Errorf("traceparent/idempotency propagation failed: trace=%q idem=%q", got.Traceparent, got.IdempotencyKey)
	}
}

func TestLoginHandler_PropagatesContextMetadata(t *testing.T) {
	var got application.LoginUserInput
	h := NewAuthHandlers(&fakeRegister{}, &fakeLogin{fn: func(_ context.Context, in application.LoginUserInput) (application.LoginUserOutput, error) {
		got = in
		return application.LoginUserOutput{UserID: "user-1", Tokens: defaultTokens}, nil
	}}, &fakeRefresh{}, &fakeLogout{})

	r := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"user@example.com","password":"Valid1!"}`))
	r.Header.Set("X-Correlation-ID", "corr-xyz")
	r.Header.Set("X-Request-ID", "req-xyz")
	r.Header.Set("traceparent", "00-aabbccddeeff00112233445566778899-0011223344556677-01")
	r.Header.Set("Idempotency-Key", "idem-xyz")
	r.RemoteAddr = "10.0.0.1:5001"

	w := httptest.NewRecorder()
	httpHandler := NewRouter(silentLogger, 5*time.Second, h)
	httpHandler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got.CorrelationID != "corr-xyz" {
		t.Errorf("correlation id = %q, want %q", got.CorrelationID, "corr-xyz")
	}
	if got.CausationID != "req-xyz" {
		t.Errorf("causation id = %q, want %q", got.CausationID, "req-xyz")
	}
	if got.Traceparent == "" || got.IdempotencyKey != "idem-xyz" {
		t.Errorf("traceparent/idempotency propagation failed: trace=%q idem=%q", got.Traceparent, got.IdempotencyKey)
	}
}

// ── RefreshHandler ────────────────────────────────────────────────────────────

func TestRefreshHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		refresh    *fakeRefresh
		wantStatus int
		wantCode   string
	}{
		{
			name:       "sucesso retorna 200 com novos tokens",
			body:       `{"refresh_token":"valid-token"}`,
			refresh:    &fakeRefresh{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "json invalido retorna 400",
			body:       `{`,
			refresh:    &fakeRefresh{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name: "token expirado retorna 401",
			body: `{"refresh_token":"expired"}`,
			refresh: &fakeRefresh{fn: func(_ context.Context, _ application.RefreshAccessTokenInput) (application.RefreshAccessTokenOutput, error) {
				return application.RefreshAccessTokenOutput{}, apperrors.Unauthorized("refresh token expired", nil)
			}},
			wantStatus: http.StatusUnauthorized,
			wantCode:   apperrors.CodeUnauthorized,
		},
		{
			name:       "executor nao configurado retorna 500",
			body:       `{"refresh_token":"t"}`,
			refresh:    nil,
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperrors.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref refreshTokenExecutor
			if tt.refresh != nil {
				ref = tt.refresh
			}
			h := NewAuthHandlers(&fakeRegister{}, &fakeLogin{}, ref, &fakeLogout{})
			r := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.RefreshHandler(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body)
			}
			if tt.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tt.wantCode)
				}
			}
		})
	}
}

// ── LogoutHandler ─────────────────────────────────────────────────────────────

func TestLogoutHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		logout     *fakeLogout
		wantStatus int
		wantCode   string
	}{
		{
			name:       "sucesso retorna 204 sem corpo",
			body:       `{"refresh_token":"valid-token"}`,
			logout:     &fakeLogout{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "json invalido retorna 400",
			body:       `]`,
			logout:     &fakeLogout{},
			wantStatus: http.StatusBadRequest,
			wantCode:   apperrors.CodeInvalidArgument,
		},
		{
			name: "token invalido retorna 401",
			body: `{"refresh_token":"bad"}`,
			logout: &fakeLogout{fn: func(_ context.Context, _ application.LogoutSessionInput) error {
				return apperrors.Unauthorized("invalid refresh token", nil)
			}},
			wantStatus: http.StatusUnauthorized,
			wantCode:   apperrors.CodeUnauthorized,
		},
		{
			name:       "executor nao configurado retorna 500",
			body:       `{"refresh_token":"t"}`,
			logout:     nil,
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperrors.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out logoutSessionExecutor
			if tt.logout != nil {
				out = tt.logout
			}
			h := NewAuthHandlers(&fakeRegister{}, &fakeLogin{}, &fakeRefresh{}, out)
			r := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.LogoutHandler(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body)
			}
			if tt.wantCode != "" {
				resp := decodeError(t, w)
				if resp.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", resp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestLogoutHandler_SuccessBodyIsEmpty(t *testing.T) {
	h := newHandlers(nil, nil, nil, nil)
	r := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(`{"refresh_token":"t"}`))
	w := httptest.NewRecorder()

	h.LogoutHandler(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("body deve ser vazio para 204, got: %s", w.Body)
	}
}
