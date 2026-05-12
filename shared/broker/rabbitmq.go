package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func New(cfg Config) (*Rabbit, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("amqp dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("create channel: %w", err)
	}

	if cfg.Prefetch > 0 {
		if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("set qos: %w", err)
		}
	}

	return &Rabbit{conn: conn, ch: ch}, nil
}

func (r *Rabbit) Channel() *amqp.Channel {
	return r.ch
}

func (r *Rabbit) Close() error {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
