package broker

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler func(ctx context.Context, d amqp.Delivery) error

type channelConsumer interface {
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

func ConsumeForever(
	ctx context.Context,
	log *slog.Logger,
	ch *amqp.Channel,
	queue string,
	consumerTag string,
	handler Handler,
) error {
	return consumeForever(ctx, log, ch, queue, consumerTag, handler)
}

func consumeForever(
	ctx context.Context,
	log *slog.Logger,
	ch channelConsumer,
	queue string,
	consumerTag string,
	handler Handler,
) error {
	msgs, err := ch.Consume(queue, consumerTag, false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := handler(ctx, d); err != nil {
				log.Error("consumer handler failed", "queue", queue, "error", err.Error())
				_ = d.Nack(false, false)
				continue
			}
			_ = d.Ack(false)
		}
	}
}
