package mongo

import (
	"context"

	"food_delivery_platform/services/auth-service/internal/domain/ports"

	mdriver "go.mongodb.org/mongo-driver/mongo"
)

type OutboxRepository struct {
	col *mdriver.Collection
}

func NewOutboxRepository(db *mdriver.Database) *OutboxRepository {
	return &OutboxRepository{col: db.Collection(outboxCollection)}
}

func (r *OutboxRepository) SavePending(ctx context.Context, msg ports.OutboxMessage) error {
	_, err := r.col.InsertOne(ctx, outboxMessageDoc{
		ID:         msg.ID,
		Exchange:   msg.Exchange,
		RoutingKey: msg.RoutingKey,
		Body:       msg.Body,
		Headers:    msg.Headers,
		CreatedAt:  msg.CreatedAt.UTC(),
		LastError:  msg.LastError,
	})
	return err
}
