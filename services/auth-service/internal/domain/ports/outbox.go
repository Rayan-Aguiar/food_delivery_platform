package ports

import (
	"context"
	"time"
)

type OutboxMessage struct {
	ID         string
	Exchange   string
	RoutingKey string
	Body       []byte
	Headers    map[string]string
	CreatedAt  time.Time
	LastError  string
}

type OutboxRepository interface {
	SavePending(ctx context.Context, msg OutboxMessage) error
}
