package application

import (
	"context"

	"food_delivery_platform/services/auth-service/internal/domain/ports"
)

type noopAuthEventPublisher struct{}

func (noopAuthEventPublisher) PublishUserRegistered(context.Context, ports.UserRegisteredEvent) error {
	return nil
}

func (noopAuthEventPublisher) PublishLoginSucceeded(context.Context, ports.LoginSucceededEvent) error {
	return nil
}
