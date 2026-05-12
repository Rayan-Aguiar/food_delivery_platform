package middleware

type contextKey string

const (
	RequestIDKey     contextKey = "request_id"
	CorrelationIDKey contextKey = "correlation_id"
)
