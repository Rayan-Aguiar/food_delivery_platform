package broker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type consumerMock struct {
	msgs      <-chan amqp.Delivery
	returnErr error
}

func (m *consumerMock) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return m.msgs, nil
}

type ackMock struct {
	ackCalls  int
	nackCalls int
}

func (m *ackMock) Ack(tag uint64, multiple bool) error {
	m.ackCalls++
	return nil
}

func (m *ackMock) Nack(tag uint64, multiple bool, requeue bool) error {
	m.nackCalls++
	return nil
}

func (m *ackMock) Reject(tag uint64, requeue bool) error {
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestConsumeForever_ConsumeError(t *testing.T) {
	err := consumeForever(context.Background(), testLogger(), &consumerMock{returnErr: errors.New("boom")}, "q", "c", func(ctx context.Context, d amqp.Delivery) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected consume error")
	}
}

func TestConsumeForever_AckOnSuccess(t *testing.T) {
	ch := make(chan amqp.Delivery, 1)
	ack := &ackMock{}
	ch <- amqp.Delivery{Acknowledger: ack, DeliveryTag: 1}
	close(ch)

	err := consumeForever(context.Background(), testLogger(), &consumerMock{msgs: ch}, "q", "c", func(ctx context.Context, d amqp.Delivery) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ack.ackCalls != 1 || ack.nackCalls != 0 {
		t.Fatalf("unexpected ack/nack calls: ack=%d nack=%d", ack.ackCalls, ack.nackCalls)
	}
}

func TestConsumeForever_NackOnHandlerError(t *testing.T) {
	ch := make(chan amqp.Delivery, 1)
	ack := &ackMock{}
	ch <- amqp.Delivery{Acknowledger: ack, DeliveryTag: 1}
	close(ch)

	err := consumeForever(context.Background(), testLogger(), &consumerMock{msgs: ch}, "q", "c", func(ctx context.Context, d amqp.Delivery) error {
		return errors.New("handler fail")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ack.nackCalls != 1 || ack.ackCalls != 0 {
		t.Fatalf("unexpected ack/nack calls: ack=%d nack=%d", ack.ackCalls, ack.nackCalls)
	}
}

func TestConsumeForever_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan amqp.Delivery)
	m := &consumerMock{msgs: c}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := consumeForever(ctx, testLogger(), m, "q", "c", func(ctx context.Context, d amqp.Delivery) error {
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got: %v", err)
	}
}
