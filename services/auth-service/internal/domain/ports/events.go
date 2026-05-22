package ports

import "context"

type AuthEventPublisher interface {
	PublishUserRegistered(ctx context.Context, in UserRegisteredEvent) error
	PublishLoginSucceeded(ctx context.Context, in LoginSucceededEvent) error
}

type UserRegisteredEvent struct {
	UserID         string
	Email          string
	RegisteredAt   string
	CorrelationID  string
	CausationID    string
	Traceparent    string
	IdempotencyKey string
}

type LoginSucceededEvent struct {
	UserID         string
	LoggedAt       string
	CorrelationID  string
	CausationID    string
	Traceparent    string
	IdempotencyKey string
}
