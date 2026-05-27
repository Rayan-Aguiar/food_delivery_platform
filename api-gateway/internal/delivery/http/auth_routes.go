package httpdelivery

import "net/http"

func RegisterAuthRoutes(mux *http.ServeMux, authProxy http.Handler) {
	if authProxy == nil {
		return
	}

	mux.Handle("POST /auth/register", authProxy)
	mux.Handle("POST /auth/login", authProxy)
	mux.Handle("POST /auth/logout", authProxy)
	mux.Handle("POST /auth/refresh", authProxy)
}
