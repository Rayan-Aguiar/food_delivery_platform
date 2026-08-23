package httpdelivery

import (
	"context"
	"net/http"
	"strings"

	apperrors "food_delivery_platform/shared/errors"
)

type tokenValidator interface {
	ValidateAccessToken(ctx context.Context, token string) (string, error)
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

// JWTAuth exige um Bearer token valido, assinado pelo auth-service, para acessar
// rotas administrativas. O contrato de claims atual do auth-service carrega
// apenas o subject autenticado (sem papel/role), entao esta camada so distingue
// autenticado (200) de nao autenticado (401) — nao ha um dado de autorizacao
// para produzir um 403 legitimo aqui.
func JWTAuth(validator tokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if validator == nil {
				writeAppError(w, r, apperrors.Internal("token validator is not configured", nil))
				return
			}

			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				writeAppError(w, r, apperrors.Unauthorized("missing bearer token", nil))
				return
			}

			if _, err := validator.ValidateAccessToken(r.Context(), token); err != nil {
				writeAppError(w, r, apperrors.Unauthorized("invalid or expired token", nil))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
