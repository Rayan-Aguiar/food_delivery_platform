package broker

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type channelPublisher interface {
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

func PublishJSON(
	ctx context.Context,
	ch *amqp.Channel,
	exchange string,
	routingKey string,
	msg any,
	headers amqp.Table,
) error {
	return publishJSON(ctx, ch, exchange, routingKey, msg, headers)
}

func publishJSON(
	ctx context.Context,
	ch channelPublisher,
	exchange string,
	routingKey string,
	msg any,
	headers amqp.Table,
) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal msg: %w", err)
	}

	pub := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
		Body:         body,
	}

	if err := ch.PublishWithContext(ctx, exchange, routingKey, false, false, pub); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}
