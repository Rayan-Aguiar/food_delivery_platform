package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/shared/broker"
	"food_delivery_platform/shared/contracts"
	"food_delivery_platform/shared/events"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultAuthExchange          = "auth.exchange"
	routingKeyUserAuthRegistered = "user.auth.registered"
	routingKeyAuthLoginSucceeded = "auth.login.succeeded"
)

type RabbitAuthEventPublisher struct {
	ch          *amqp.Channel
	exchange    string
	serviceName string
	log         *slog.Logger
	outbox      ports.OutboxRepository
	retry       broker.RetryPolicy
	publishFn   func(ctx context.Context, ch *amqp.Channel, exchange, routingKey string, msg any, headers amqp.Table) error
	waitFn      func(ctx context.Context, d time.Duration) error
}

func NewRabbitAuthEventPublisher(
	ch *amqp.Channel,
	exchange string,
	serviceName string,
	log *slog.Logger,
	retry broker.RetryPolicy,
	outbox ...ports.OutboxRepository,
) *RabbitAuthEventPublisher {
	if exchange == "" {
		exchange = defaultAuthExchange
	}
	if retry.MaxAttempts <= 0 {
		retry.MaxAttempts = 1
	}
	if retry.BaseDelay <= 0 {
		retry.BaseDelay = 100 * time.Millisecond
	}
	if retry.MaxDelay <= 0 {
		retry.MaxDelay = 2 * time.Second
	}

	var outboxRepo ports.OutboxRepository
	if len(outbox) > 0 {
		outboxRepo = outbox[0]
	}

	return &RabbitAuthEventPublisher{
		ch:          ch,
		exchange:    exchange,
		serviceName: serviceName,
		log:         log,
		outbox:      outboxRepo,
		retry:       retry,
		publishFn:   broker.PublishJSON,
		waitFn:      waitWithContext,
	}
}

func waitWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type userRegisteredPayload struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	RegisteredAt string `json:"registered_at"`
}

type loginSucceededPayload struct {
	UserID   string `json:"user_id"`
	LoggedAt string `json:"logged_at"`
}

func (p *RabbitAuthEventPublisher) PublishUserRegistered(ctx context.Context, in ports.UserRegisteredEvent) error {
	payload := userRegisteredPayload{
		UserID:       in.UserID,
		Email:        in.Email,
		RegisteredAt: in.RegisteredAt,
	}
	env := events.NewEnvelope(
		events.UserAuthRegistered,
		1,
		p.serviceName,
		in.CorrelationID,
		in.CausationID,
		in.IdempotencyKey,
		in.Traceparent,
		payload,
	)
	if err := events.ValidateMeta(env.Meta); err != nil {
		return fmt.Errorf("invalid event meta: %w", err)
	}

	headers := buildEventHeaders(env.Meta)
	return p.publishWithRetry(ctx, routingKeyUserAuthRegistered, env, headers)
}

func (p *RabbitAuthEventPublisher) PublishLoginSucceeded(ctx context.Context, in ports.LoginSucceededEvent) error {
	payload := loginSucceededPayload{
		UserID:   in.UserID,
		LoggedAt: in.LoggedAt,
	}
	env := events.NewEnvelope(
		events.AuthLoginSucceeded,
		1,
		p.serviceName,
		in.CorrelationID,
		in.CausationID,
		in.IdempotencyKey,
		in.Traceparent,
		payload,
	)
	if err := events.ValidateMeta(env.Meta); err != nil {
		return fmt.Errorf("invalid event meta: %w", err)
	}

	headers := buildEventHeaders(env.Meta)
	return p.publishWithRetry(ctx, routingKeyAuthLoginSucceeded, env, headers)
}

func buildEventHeaders(meta contracts.EventMeta) amqp.Table {
	headers := amqp.Table{
		"event_id":       meta.EventID,
		"event_type":     meta.EventType,
		"event_version":  int32(meta.EventVersion),
		"occurred_at":    meta.OccurredAt,
		"producer":       meta.Producer,
		"correlation_id": meta.CorrelationID,
	}
	if meta.CausationID != "" {
		headers["causation_id"] = meta.CausationID
	}
	if meta.IdempotencyKey != "" {
		headers["idempotency_key"] = meta.IdempotencyKey
	}
	if meta.Traceparent != "" {
		headers["traceparent"] = meta.Traceparent
	}
	return headers
}

func (p *RabbitAuthEventPublisher) publishWithRetry(ctx context.Context, routingKey string, msg any, headers amqp.Table) error {
	ctx, span := otel.Tracer("auth-service.messaging").Start(ctx, "rabbitmq.publish", trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", p.exchange),
		attribute.String("messaging.rabbitmq.routing_key", routingKey),
		attribute.String("event.id", headerString(headers, "event_id")),
		attribute.String("event.type", headerString(headers, "event_type")),
	)

	injectTraceparent(ctx, headers)

	var lastErr error

	for attempt := 1; attempt <= p.retry.MaxAttempts; attempt++ {
		headers["attempt"] = int32(attempt)
		err := p.publishFn(ctx, p.ch, p.exchange, routingKey, msg, headers)
		if err == nil {
			span.SetStatus(codes.Ok, "publish succeeded")
			return nil
		}
		span.RecordError(err)
		if p.log != nil {
			p.log.Warn("auth event publish failed",
				"exchange", p.exchange,
				"routing_key", routingKey,
				"attempt", attempt,
				"correlation_id", headerString(headers, "correlation_id"),
				"event_id", headerString(headers, "event_id"),
				"event_type", headerString(headers, "event_type"),
				"error", err.Error(),
			)
		}
		lastErr = err
		if attempt == p.retry.MaxAttempts {
			break
		}

		delay := p.retry.NextDelay(attempt)
		if err := p.waitFn(ctx, delay); err != nil {
			return err
		}
	}

	if p.outbox != nil {
		body, marshalErr := json.Marshal(msg)
		if marshalErr != nil {
			if p.log != nil {
				p.log.Error("failed to marshal event for outbox fallback",
					"routing_key", routingKey,
					"event_id", headerString(headers, "event_id"),
					"error", marshalErr.Error(),
				)
			}
		} else {
			saveErr := p.outbox.SavePending(ctx, ports.OutboxMessage{
				ID:         headerString(headers, "event_id"),
				Exchange:   p.exchange,
				RoutingKey: routingKey,
				Body:       body,
				Headers:    normalizeHeaders(headers),
				CreatedAt:  time.Now().UTC(),
				LastError:  lastErr.Error(),
			})
			if p.log != nil {
				if saveErr != nil {
					p.log.Error("failed to persist event in outbox fallback",
						"routing_key", routingKey,
						"event_id", headerString(headers, "event_id"),
						"error", saveErr.Error(),
					)
				} else {
					p.log.Info("event stored in outbox fallback",
						"routing_key", routingKey,
						"event_id", headerString(headers, "event_id"),
					)
				}
			}
		}
	}

	if lastErr != nil {
		span.SetStatus(codes.Error, lastErr.Error())
	}

	return fmt.Errorf("publish after retries: %w", lastErr)
}

func injectTraceparent(ctx context.Context, headers amqp.Table) {
	if headerString(headers, "traceparent") != "" {
		return
	}
	carrier := amqpHeaderCarrier(headers)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

type amqpHeaderCarrier amqp.Table

func (c amqpHeaderCarrier) Get(key string) string {
	v, ok := c[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return fmt.Sprint(v)
}

func (c amqpHeaderCarrier) Set(key, value string) {
	c[key] = value
}

func (c amqpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

var _ propagation.TextMapCarrier = amqpHeaderCarrier{}

func headerString(headers amqp.Table, key string) string {
	v, ok := headers[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return fmt.Sprint(v)
}

func normalizeHeaders(headers amqp.Table) map[string]string {
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		out[k] = fmt.Sprint(v)
	}
	return out
}
