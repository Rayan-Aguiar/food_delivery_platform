package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"food_delivery_platform/services/auth-service/internal/application"
	apperrors "food_delivery_platform/shared/errors"
	"food_delivery_platform/shared/middleware"
	"food_delivery_platform/shared/utils"
)

type registerUserExecutor interface {
	Execute(ctx context.Context, input application.RegisterUserInput) (application.RegisterUserOutput, error)
}

type loginUserExecutor interface {
	Execute(ctx context.Context, input application.LoginUserInput) (application.LoginUserOutput, error)
}

type refreshTokenExecutor interface {
	Execute(ctx context.Context, input application.RefreshAccessTokenInput) (application.RefreshAccessTokenOutput, error)
}

type logoutSessionExecutor interface {
	Execute(ctx context.Context, input application.LogoutSessionInput) error
}

type AuthHandlers struct {
	register registerUserExecutor
	login    loginUserExecutor
	refresh  refreshTokenExecutor
	logout   logoutSessionExecutor
}

func NewAuthHandlers(
	register registerUserExecutor,
	login loginUserExecutor,
	refresh refreshTokenExecutor,
	logout logoutSessionExecutor,
) *AuthHandlers {
	return &AuthHandlers{register: register, login: login, refresh: refresh, logout: logout}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID       string `json:"user_id"`
	CredentialID string `json:"credential_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.register == nil {
		writeAppError(w, r, apperrors.Internal("register use case is not configured", nil))
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAppError(w, r, apperrors.InvalidArgument("invalid request body", err))
		return
	}

	out, err := h.register.Execute(r.Context(), application.RegisterUserInput{
		Email:          req.Email,
		Password:       req.Password,
		UserAgent:      r.UserAgent(),
		IPAddress:      extractClientIP(r.RemoteAddr),
		CorrelationID:  utils.StringFromContext(r.Context(), middleware.CorrelationIDKey),
		CausationID:    utils.StringFromContext(r.Context(), middleware.RequestIDKey),
		Traceparent:    r.Header.Get("traceparent"),
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusCreated, registerResponse{
		UserID:       out.UserID,
		CredentialID: out.CredentialID,
		AccessToken:  out.Tokens.AccessToken,
		RefreshToken: out.Tokens.RefreshToken,
		TokenType:    out.Tokens.TokenType,
		ExpiresIn:    out.Tokens.ExpiresIn,
	})
}

func (h *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.login == nil {
		writeAppError(w, r, apperrors.Internal("login use case is not configured", nil))
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAppError(w, r, apperrors.InvalidArgument("invalid request body", err))
		return
	}

	out, err := h.login.Execute(r.Context(), application.LoginUserInput{
		Email:          req.Email,
		Password:       req.Password,
		UserAgent:      r.UserAgent(),
		IPAddress:      extractClientIP(r.RemoteAddr),
		CorrelationID:  utils.StringFromContext(r.Context(), middleware.CorrelationIDKey),
		CausationID:    utils.StringFromContext(r.Context(), middleware.RequestIDKey),
		Traceparent:    r.Header.Get("traceparent"),
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, tokenResponse{
		UserID:       out.UserID,
		AccessToken:  out.Tokens.AccessToken,
		RefreshToken: out.Tokens.RefreshToken,
		TokenType:    out.Tokens.TokenType,
		ExpiresIn:    out.Tokens.ExpiresIn,
	})
}

func (h *AuthHandlers) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.refresh == nil {
		writeAppError(w, r, apperrors.Internal("refresh use case is not configured", nil))
		return
	}

	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAppError(w, r, apperrors.InvalidArgument("invalid request body", err))
		return
	}

	out, err := h.refresh.Execute(r.Context(), application.RefreshAccessTokenInput{RefreshToken: req.RefreshToken})
	if err != nil {
		writeAppError(w, r, err)
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, tokenResponse{
		UserID:       out.UserID,
		AccessToken:  out.Tokens.AccessToken,
		RefreshToken: out.Tokens.RefreshToken,
		TokenType:    out.Tokens.TokenType,
		ExpiresIn:    out.Tokens.ExpiresIn,
	})
}

func (h *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.logout == nil {
		writeAppError(w, r, apperrors.Internal("logout use case is not configured", nil))
		return
	}

	var req logoutRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAppError(w, r, apperrors.InvalidArgument("invalid request body", err))
		return
	}

	if err := h.logout.Execute(r.Context(), application.LogoutSessionInput{RefreshToken: req.RefreshToken}); err != nil {
		writeAppError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, dest any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func extractClientIP(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}
	return host
}
