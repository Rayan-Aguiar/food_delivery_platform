package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const CorrelationIDHeader = "X-Correlation-ID"

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(CorrelationIDHeader)
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
		w.Header().Set(CorrelationIDHeader, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
